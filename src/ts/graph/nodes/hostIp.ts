import { Port } from "../../api";
import { LinkCollection } from "../../utils/linkCollection";
import { AnyNode } from "../types";
import { BaseNode } from "./baseNode";

export interface HostIPModel {
  ID: string;
}

export class HostIPNode extends BaseNode<HostIPModel> {

  public override get icon(): string {
    return "\uf390";
  }

  public override get type(): string {
    return "Host IP";
  }
}

export class HostIPLinks extends LinkCollection<Port, HostIPModel, HostIPNode> {

  protected override mapModel({ HostIp }: Port): HostIPModel {
    return { ID: HostIp };
  }

  protected override buildNode(model: HostIPModel): HostIPNode {
    return new HostIPNode(model);
  }

  protected override isTargetNode(node: AnyNode): node is HostIPNode {
    return node instanceof HostIPNode;
  }
}
