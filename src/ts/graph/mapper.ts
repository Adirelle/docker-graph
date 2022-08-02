import { Container, Mount, Network, Port, shortIDOrName } from "../api";
import { NodeModel } from "../models";

function* iterMap<T, U>(inputs: Iterable<T>, f: (input: T) => U): Iterable<U> {
  for (const input of inputs) {
    yield f(input);
  }
}

function* iterFlatMap<T, U>(inputs: Iterable<T>, f: (input: T) => Iterable<U>): Iterable<U> {
  for (const input of inputs) {
    for (const output of f(input)) {
      yield output;
    }
  }
}

export class Mapper {

  public statusColors: { [status: string]: string; } = {
    running: "#0C0",
    exited: "#999",
  };

  public map(ctn: Container): NodeModel {
    return {
      type: "container",
      id: ctn.ID,
      label: shortIDOrName(ctn.Name || ctn.ID),
      tooltip: `container: ${ctn.Name}<br/>id: ${ctn.ID}<br/>status: ${ctn.Status}`,
      color: this.statusColors[ctn.Status] || 'black',
      links: [
        this.mapImage(ctn.Image),
        ...iterFlatMap(Object.values(ctn.Networks || {}), n => this.mapNetwork(n)),
        ...iterFlatMap(Object.values(ctn.Mounts || []), m => this.mapMount(m)),
        ...iterMap(Object.entries(ctn.Ports || {}), p => this.mapPort(p, ctn))
      ]
    };
  }

  private mapImage(img: string): NodeModel {
    return {
      type: "image",
      id: `img:${img}`,
      label: img
    };
  }

  private *mapNetwork(net: Network): Iterable<NodeModel> {
    if (net.ID) {
      yield {
        type: "network",
        id: net.ID,
        label: shortIDOrName(net.Name || net.ID)
      };
    }
  }

  private *mapMount(mnt: Mount): Iterable<NodeModel> {
    switch (mnt.Type) {
      case "volume":
        yield {
          type: "volume",
          id: mnt.Name,
          label: shortIDOrName(mnt.Name),
        };
        break;
      case "bind":
        yield {
          type: "bindMount",
          id: mnt.Source,
          label: mnt.Source
        };
        break;
    }
  }

  private mapPort([name, port]: [string, Port], ctn: Container): NodeModel {
    return {
      type: "port",
      id: `${ctn.ID}:${name}`,
      label: name,
      links: [
        {
          type: "hostIP",
          id: `IP:${port.HostIp}`,
          label: port.HostIp,
        }
      ]
    };
  }

}
