import { Port } from "../../api";
import { LinkCollection } from "../../utils/linkCollection";
import { AnyNode, Context } from "../types";
import { BaseNode } from "./baseNode";
import { HostIPLinks } from "./hostIp";

export interface PortModel extends Port {
  ID: string;
  Name: string;
}

export class PortNode extends BaseNode<PortModel> {
  private hostIPs = new HostIPLinks(this);

  public override get icon(): string {
    return "\uf796";
  }

  public override get type(): string {
    return "port";
  }

  public override updateFrom(model: PortModel, ctx: Context): boolean {
    let changed = super.updateFrom(model, ctx);

    changed = this.hostIPs.update([model], ctx);

    return changed;
  }
}

export class PortLinks extends LinkCollection<[string, Port], PortModel, PortNode> {

  protected override mapModel([name, port]: [string, Port]): PortModel {
    return { ID: `${this.source.id}:${name}`, Name: name, ...port };
  }

  protected override buildNode(model: PortModel): PortNode {
    return new PortNode(model);
  }

  protected override isTargetNode(node: AnyNode): node is PortNode {
    return node instanceof PortNode;
  }
}
