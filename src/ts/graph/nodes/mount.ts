import { BaseNode } from "./baseNode";

export interface MountModel {
  ID: string;
  Name: string;
  Destination: string;
  ReadWrite: boolean;
}

export abstract class MountNode<T extends MountModel> extends BaseNode<T> {
  public override get icon(): string {
    return "x";
  }

  public override get type(): string {
    return "mount";
  }

  protected override tooltipLines(model: T): [string, string][] {
    return [
      ... super.tooltipLines(model),
      ["dest", model.Destination],
      ["writable", model.ReadWrite ? "yes" : "no"]
    ];
  }
}
