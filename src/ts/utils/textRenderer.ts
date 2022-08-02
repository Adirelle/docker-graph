type Alignment = "left" | "center" | "right" | "start" | "end";
type Baseline =
  | "top"
  | "hanging"
  | "middle"
  | "alphabetic"
  | "ideographic"
  | "bottom";
type Weight = number | "normal" | "bold" | "lighter" | "bolder";

export interface TextRenderOptions {
  size: number;
  font: string;
  weight: Weight;
  alignment: Alignment;
  baseline: Baseline;
}

class _DefaultTextOptions implements TextRenderOptions {
  public size = 6;
  public font = "sans-serif";
  public weight: Weight = "normal";
  public alignment: Alignment = "center";
  public baseline: Baseline = "middle";
}

export const DefaultTextOptions = new _DefaultTextOptions();

export class TextRenderer extends _DefaultTextOptions {

  public constructor(options: Partial<TextRenderOptions> = {}) {
    super();
    Object.assign(this, options);
  }

  protected actualSize(_scale: number): number {
    return this.size;
  }

  protected prepare(ctx: CanvasRenderingContext2D, scale: number): void {
    ctx.font = `${this.weight} ${this.actualSize(scale)}px "${this.font}"`;
    ctx.textAlign = this.alignment;
    ctx.textBaseline = this.baseline;
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

export class ConstantSizeTextRenderer extends TextRenderer {

  protected override actualSize(scale: number): number {
    return this.size / scale;
  }

}

export const labelTextRenderer = new TextRenderer();

export const iconRenderer = new TextRenderer({ font: `FontAwesome`, size: 12 });

