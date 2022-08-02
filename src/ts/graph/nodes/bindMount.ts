import { Mount } from "../../api";
import { LinkCollection } from "../../utils/linkCollection";
import { AnyNode } from "../types";
import { MountModel, MountNode } from "./mount";

export interface BindMount extends MountModel {
}

export class BindMountNode extends MountNode<BindMount> {
  public override get icon(): string {
    return "\uf07b";
  }

  public override get type(): string {
    return "Bind mount";
  }
}

export class BindMountLinks extends LinkCollection<Mount, BindMount, BindMountNode> {

  protected override accept(item: Mount): boolean {
    return item.Type == "bind";
  }

  protected override mapModel(model: Mount): BindMount {
    console.debug("BindMount", model);
    return {
      ID: model.Source,
      Name: model.Source,
      Destination: model.Destination,
      ReadWrite: model.ReadWrite,
    };
  }

  protected override buildNode(model: BindMount): BindMountNode {
    return new BindMountNode(model);
  }

  protected override isTargetNode(node: AnyNode): node is BindMountNode {
    return node instanceof BindMountNode;
  }
}
