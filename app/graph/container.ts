import { Container } from "../api";
import { BaseNode } from "./baseNode";
import { NetworkNode } from "./network";
import { TextRenderer } from "./render";
import { Context } from "./types";

const labelRenderer = new TextRenderer({});

const statusColors: { [status: string]: string } = {
  running: "#0C0",
  exited: "#CCC",
};

export class ContainerNode extends BaseNode<Container> {
  public readonly image: string;
  public status: string = "unknown";
  public width: number = 5;
  public height: number = 5;

  public get name(): string {
    return `image: ${this.image}<br/>status: ${this.status}`;
  }

  public constructor(data: Container) {
    super(data);
    this.label = data.Service || data.Name;
    this.image = data.Image;
  }

  public override updateFrom(ctn: Container, ctx: Context): boolean {
    let dirty = super.updateFrom(ctn, ctx);
    if (this.status != ctn.Status) {
      this.status = ctn.Status;
      dirty = true;
    }

    const nets = ctn.Networks || {};
    const netNodes = new Set<NetworkNode>();
    for (const [_, network] of Object.entries(nets)) {
      if (!network.ID) continue;
      const node = ctx.getOrCreateNode(network.ID, NetworkNode, network);
      ctx.getOrCreateLink(this, node);
      netNodes.add(node);
    }
    for (const link of ctx.listLinksFrom(this)) {
      if (link.target instanceof NetworkNode && !netNodes.has(link.target)) {
        ctx.removeLink(link);
      }
    }

    return dirty;
  }

  public override render(ctx: CanvasRenderingContext2D, scale: number): void {
    const { x, y, label, status } = this;
    if (!x || !y || !label) return;

    let [w, h] = labelRenderer.measure(ctx, scale, label);
    w += 3;
    h += 3;
    this.width = w;
    this.height = h;

    ctx.strokeStyle = "black";
    ctx.lineWidth = 0.5;
    ctx.fillStyle = statusColors[status] || "white";
    ctx.beginPath();
    ctx.rect(x - w / 2, y - h / 2, w, h);
    ctx.fill();
    ctx.stroke();

    labelRenderer.render(ctx, scale, label, x, y);
  }

  public override paintInteractionArea(
    color: string,
    ctx: CanvasRenderingContext2D,
    _: number
  ): void {
    const { x, y, width, height } = this;
    if (!x || !y) return;

    ctx.fillStyle = color;
    ctx.fillRect(x - width / 2, y - height / 2, width, height);
  }
}
