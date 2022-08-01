type TextAlign = "left" | "center" | "right" | "start" | "end";
type TextBaseline =
  | "top"
  | "hanging"
  | "middle"
  | "alphabetic"
  | "ideographic"
  | "bottom";

export interface TextRenderOptions {
  color: string;
  baseSize: number;
  font: string;
  align: TextAlign;
  baseline: TextBaseline;
  dynamic: boolean;
}

export class TextRenderer implements TextRenderOptions {
  color: string;
  baseSize: number;
  font: string;
  align: TextAlign;
  baseline: TextBaseline;
  dynamic: boolean;

  public constructor({
    color,
    baseSize,
    font,
    align,
    baseline,
    dynamic,
  }: Partial<TextRenderOptions> = {}) {
    this.color = color || "black";
    this.baseSize = baseSize || 12;
    this.font = font || "sans-serif";
    this.align = align || "center";
    this.baseline = baseline || "middle";
    this.dynamic = dynamic !== undefined ? dynamic : true;
  }

  protected prepare(ctx: CanvasRenderingContext2D, scale: number): void {
    if (this.dynamic) {
      ctx.font = `${this.baseSize / scale}pt ${this.font}`;
    } else {
      ctx.font = `${this.baseSize}pt ${this.font}`;
    }
    ctx.textAlign = this.align;
    ctx.textBaseline = this.baseline;
    ctx.fillStyle = this.color;
  }

  public measure(
    ctx: CanvasRenderingContext2D,
    scale: number,
    text: string
  ): [number, number] {
    this.prepare(ctx, scale);
    const metrics = ctx.measureText(text);
    return [
      metrics.actualBoundingBoxRight + metrics.actualBoundingBoxLeft,
      metrics.actualBoundingBoxAscent + metrics.actualBoundingBoxDescent,
    ];
  }

  public render(
    ctx: CanvasRenderingContext2D,
    scale: number,
    text: string,
    x: number,
    y: number
  ): void {
    this.prepare(ctx, scale);
    ctx.fillText(text, x, y);
  }
}
