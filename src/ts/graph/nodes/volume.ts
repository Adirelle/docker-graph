import { Mount } from "../../api";
import { LinkCollection } from "../../utils/linkCollection";
import { AnyNode } from "../types";
import { MountModel, MountNode } from "./mount";

export interface Volume extends MountModel {
}

export class VolumeNode extends MountNode<Volume> {
  public override get icon(): string {
    return "\uf1c0";
  }

  public override get type(): string {
    return "volume";
  }
}

export class VolumeLinks extends LinkCollection<Mount, Volume, VolumeNode> {

  protected override accept(item: Mount): boolean {
    return item.Type == "volume";
  }

  protected override mapModel(mount: Mount): Volume {
    console.debug("volume", mount);
    return { ID: mount.Source, Name: mount.Name, Destination: mount.Destination, ReadWrite: mount.ReadWrite };
  }

  protected override buildNode(model: Volume): VolumeNode {
    return new VolumeNode(model);
  }

  protected override isTargetNode(node: AnyNode): node is VolumeNode {
    return node instanceof VolumeNode;
  }
}
