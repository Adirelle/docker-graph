import { shortIDOrName } from "../../api";
import { iconRenderer, labelTextRenderer } from "../../utils/textRenderer";
import { Context, Node } from "../types";

export interface NodeModel {
  ID: string;
  Name?: string;
}

export interface NodeProps {
  label: string;
  tooltip: string;
  color: String;
}

export abstract class BaseNode<T extends NodeModel> implements Node<T>, NodeProps {
  public id: string;

  public x?: number;
  public y?: number;
  public vx?: number;
  public vy?: number;
  public fx?: number;
  public fy?: number;
  public width?: number;
  public height?: number;

  public label: string;
  public color = "black";
  public tooltip: string = "";

  public constructor(model: T) {
    this.id = model.ID;
    this.label = model.Name || shortIDOrName(model.ID);
  }

  public get icon(): string {
    return "x";
  }

  public get type(): string {
    return "?";
  }

  public updateFrom(model: T, _ctx: Context): boolean {
    let changed = false;

    const props = this.mapModel(model);
    const thisProps = this as { [key: string]: any; };
    for (const [key, value] of Object.entries(props)) {
      if (thisProps[key] !== value) {
        thisProps[key] = value;
        changed = true;
      }
    }

    return changed;
  }

  public isVisible(): boolean {
    return true;
  }

  public render(ctx: CanvasRenderingContext2D, scale: number): void {
    let { x, y } = this;
    if (!x || !y) return;

    ctx.fillStyle = this.color;
    [this.width, this.height] = this.renderIcon(ctx, scale, x, y);
    y += this.width;
    this.renderLabel(ctx, scale, x, y);
  }

  public paintInteractionArea(
    color: string,
    ctx: CanvasRenderingContext2D,
    _scale: number
  ): void {
    const { x, y, width, height } = this;
    if (!x || !y || !width || !height) return;

    ctx.fillStyle = color;
    ctx.fillRect(x - width / 2, y - height / 2, width, height);
  }

  protected mapModel(model: T): NodeProps {
    const tooltip = [[this.type, this.label], ... this.tooltipLines(model)];
    return {
      label: shortIDOrName(model.Name || model.ID),
      color: "black",
      tooltip: tooltip.map(([k, v]) => `${k}: ${v}`).join("<br/>")
    };
  }

  protected tooltipLines(model: T): Array<[string, string]> {
    return [];
  }

  protected renderIcon(ctx: CanvasRenderingContext2D, scale: number, x: number, y: number): [number, number] {
    iconRenderer.render(ctx, scale, this.icon, x, y);
    return iconRenderer.measure(ctx, scale, this.icon);
  }

  protected renderLabel(ctx: CanvasRenderingContext2D, scale: number, x: number, y: number): void {
    labelTextRenderer.render(ctx, scale, this.label, x, y);
  }
}
