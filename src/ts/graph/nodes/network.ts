import { Network } from "../../api";
import { LinkCollection } from "../../utils/linkCollection";
import { AnyNode } from "../types";
import { BaseNode } from "./baseNode";

export class NetworkNode extends BaseNode<Network> {
  public override get icon(): string {
    return "\uf6ff";
  }

  public override get type(): string {
    return "network";
  }
}

export class NetworkLinks extends LinkCollection<Network, Network, NetworkNode> {

  protected override accept(item: Network): boolean {
    return item.ID != "";
  }

  protected override mapModel(item: Network): Network {
    return item;
  }

  protected override buildNode(model: Network): NetworkNode {
    return new NetworkNode(model);
  }

  protected override isTargetNode(node: AnyNode): node is NetworkNode {
    return node instanceof NetworkNode;
  }
}

