import { Container } from "../../api";
import { Context } from "../types";
import { BaseNode, NodeProps } from "./baseNode";
import { BindMountLinks } from "./bindMount";
import { ImageLinks } from "./image";
import { NetworkLinks } from "./network";
import { PortLinks } from "./port";
import { VolumeLinks } from "./volume";

const statusColors: { [status: string]: string; } = {
  running: "#0C0",
  exited: "#999",
};

export class ContainerNode extends BaseNode<Container> {
  private networks = new NetworkLinks(this);
  private ports = new PortLinks(this);
  private volumes = new VolumeLinks(this);
  private bindMounts = new BindMountLinks(this);
  private images = new ImageLinks(this);

  public override get type(): string {
    return "container";
  }

  public override get icon(): string {
    return "\uf395";
  }

  public override updateFrom(ctn: Container, ctx: Context): boolean {
    let changed = super.updateFrom(ctn, ctx);

    const networks = Object.values(ctn.Networks || {});
    changed = this.networks.update(networks, ctx) || changed;

    const ports = Object.entries(ctn.Ports || {});
    changed = this.ports.update(ports, ctx) || changed;

    const mounts = ctn.Mounts || [];
    changed = this.volumes.update(mounts, ctx) || changed;
    changed = this.bindMounts.update(mounts, ctx) || changed;

    changed = this.images.update([ctn.Image], ctx);

    return changed;
  }

  protected override mapModel(model: Container): NodeProps {
    return {
      ... super.mapModel(model),
      color: statusColors[model.Status] || "black",
      label: model.Service || model.Name
    };
  }

  protected override tooltipLines(model: Container): Array<[string, string]> {
    return [
      ["status", model.Status]
    ];
  }
}

