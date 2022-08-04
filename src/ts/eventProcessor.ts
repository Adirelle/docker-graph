import { Container, Event } from "./api";
import { NodeModel } from "./models";
import { parseImage, shortName, shortPath } from "./utils";

export type NodeUpdateFunc = (node: NodeModel, updater: Updater) => void;

export interface Updater {
  updateNode(id: string, update: NodeUpdateFunc): void;
  removeNode(id: string): void;
  updateLink(sourceID: string, targetID: string, update: NodeUpdateFunc): void;
  tidy(): void;
}

export class EventProcessor {

  public constructor(
    private readonly updaterFactory: () => Updater
  ) { }

  public process(event: Event): boolean {
    if (event.TargetType != "container") {
      return false;
    }
    const updater = this.updaterFactory();
    if (event.Type == "removed") {
      updater.removeNode(event.TargetID);
    } else {
      updater.updateNode(event.TargetID, (n, u) => this.updateContainer(n, event.Details, u));
    }
    updater.tidy();
    return true;
  }

  private updateContainer(node: NodeModel, ctn: Container, updater: Updater): void {
    node.type = "container";
    node.label = shortName(ctn.Name || ctn.ID, ctn.Project);
    node.tooltip = makeTooltip(
      "container", ctn.Name,
      "id", ctn.ID,
      "status", ctn.Status,
      "project", ctn.Project?.Name || "none"
    );
    switch (ctn.Status) {
      case 'running':
        node.color = '#070';
        console.log("healty?", ctn.Healty);
        break;
      case 'exited':
        node.color = '#888';
        break;
      default:
        delete node.color;
    }

    const imageID = `img:${ctn.Image}`;
    updater.updateLink(ctn.ID, imageID, (node) => {
      node.type = "image";
      const { Name, Registry, Tag } = parseImage(ctn.Image);
      node.label = shortName(Name, ctn.Project);
      node.tooltip = makeTooltip(
        "image", ctn.Image,
        "registry", Registry,
        "name", Name,
        "tag", Tag
      );
    });

    for (const net of Object.values(ctn.Networks || {})) {
      if (net.ID == "") continue;
      updater.updateLink(ctn.ID, net.ID, (node) => {
        node.type = "network";
        node.label = shortName(net.Name || net.ID, ctn.Project);
      });
    }

    for (const mount of (ctn.Mounts || [])) {
      switch (mount.Type) {
        case "bind":
          updater.updateLink(ctn.ID, mount.Source, (node) => {
            node.type = "bindMount";
            node.label = shortPath(mount.Source, ctn.Project);
            node.tooltip = makeTooltip(
              "bind mount", mount.Source,
              "source", mount.Source,
              "destination", mount.Destination,
              "writable?", mount.ReadWrite ? "yes" : "no"
            );
          });
          break;
        case "volume":
          updater.updateLink(ctn.ID, mount.Source, (node) => {
            node.type = "volume";
            node.label = shortName(mount.Name, ctn.Project);
            node.tooltip = makeTooltip(
              "volume", mount.Name,
              "source", mount.Source,
              "destination", mount.Destination,
              "writable?", mount.ReadWrite ? "yes" : "no"
            );
          });
          break;
      }
    }

    for (const [inner, binding] of Object.entries(ctn.Ports || {})) {
      const id = `${ctn.ID}:${inner}`;
      updater.updateLink(ctn.ID, id, (node) => {
        node.type = "port",
          node.label = inner;
        updater.updateLink(id, `IP:${binding.HostIp}`, (node) => {
          node.type = "hostIP";
          node.label = binding.HostIp;
        });
      });
    }
  }
}

function makeTooltip(...parts: string[]): string {
  const lines = [];
  for (let i = 0, l = parts.length; i < l; i += 2) {
    lines.push(`${parts[i]}: ${parts[i + 1]}`);
  }
  return lines.join("<br/>");
}
