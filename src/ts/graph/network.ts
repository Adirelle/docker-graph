import { Network } from "../api";
import { BaseNode } from "./baseNode";
import { TextRenderer } from "./render";

const labelRenderer = new TextRenderer({ baseline: "top" });

const iconRenderer = new TextRenderer({
  font: "Font Awesome 6 Free",
  baseSize: 32,
  dynamic: false,
});

export class NetworkNode extends BaseNode<Network> {
  public constructor(net: Network) {
    super(net);
    console.debug("network", net);
    this.label = net.Name;
  }

  public override render(ctx: CanvasRenderingContext2D, scale: number): void {
    const { x, y, label } = this;
    if (!x || !y) return;

    iconRenderer.render(ctx, scale, "\uf6ff", x, y);
    if (label) {
      labelRenderer.render(ctx, scale, label, x, y + 3);
    }
  }

  public override paintInteractionArea(
    color: string,
    ctx: CanvasRenderingContext2D,
    _: number
  ): void {
    const { x, y } = this;
    if (!x || !y) return;

    ctx.fillStyle = color;
    ctx.fillRect(x - 5, y - 5, 10, 10);
  }
}
