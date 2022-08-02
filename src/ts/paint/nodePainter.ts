import { NodeModel, NodeType } from "../models";
import { TextRenderer } from "./text";

type IconMap = Record<NodeType, string>;

const typeIcons: IconMap = {
  container: '\uf395',
  network: '\uf6ff',
  hostIP: '\uf390',
  image: '\uf03e',
  bindMount: '\uf07b',
  port: '\uf796',
  volume: '\uf1c0',
};

export class NodePainter {

  public constructor(
    private readonly labelRenderer: TextRenderer,
    private readonly iconRenderer: TextRenderer
  ) { }

  public paint(node: NodeModel, ctx: CanvasRenderingContext2D, scale: number): void {
    let { x, y, type, color, label } = node;
    if (!x || !y) return;

    ctx.fillStyle = color || "black";
    const icon = typeIcons[type] || 'x';
    this.iconRenderer.render(ctx, scale, icon, x, y);
    [node.width, node.height] = this.iconRenderer.measure(ctx, scale, icon);

    if (label) {
      y += node.height;
      this.labelRenderer.render(ctx, scale, label, x, y);
    }
  }

  public paintInteractionArea(
    { x, y, width, height }: NodeModel,
    color: string,
    ctx: CanvasRenderingContext2D
  ): void {
    if (!x || !y || !width || !height) return;

    ctx.fillStyle = color;
    ctx.fillRect(x - width / 2, y - height / 2, width, height);
  }
}
