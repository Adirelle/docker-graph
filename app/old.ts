import forceGraph, { LinkObject, NodeObject } from 'force-graph';

import { Container, ContainerList, Network, Volume } from './api';

import './style.css';

async function jsonFetch<T>(url: string): Promise<T> {
  const resp = await fetch(url);
  return resp.json();
}

type Node = NodeObject&{ type: string; name: string; label?: string;};
type Link = LinkObject;

async function bootstrap() {


  const containerList = await jsonFetch<ContainerList>("/api/containers")
  const containers = await Promise.all(containerList.Containers.map(id => jsonFetch<Container>(`/api/containers/${id}`)));

  console.log("containers", containers);

  const networkIds = new Set<string>();
  const volumeIds = new Set<string>();

  const nodes: Node[] = [];
  const links: Link[] = [];

  for (const { id, name, service, networks, mounts } of containers) {
    nodes.push({ id, name, type: "container", label: service || name });
    if (networks) {
      for (const networkID of networks) {
        links.push({ source: id, target: networkID });
        networkIds.add(networkID);
      }
    }
    if (mounts) {
      for (const { type, name, source, destination } of mounts) {
        if (type == "bind") {
          nodes.push({ id: source, name: destination, type: "bind", label: source })
          links.push({ source: id, target: source });
        } else if (type == "volume") {
          volumeIds.add(name);
          links.push({ source: id, target: name });
        }
      }
    }
  }

  const networks = await Promise.all(
    Array.from(networkIds)
      .map(id => jsonFetch<Network>(`/api/networks/${id}`))
  )

  console.log("networks", networks);

  for (const { id, name } of networks) {
    nodes.push({ id, name, type: "network" });
  }

  const volumes = await Promise.all(
    Array.from(volumeIds)
      .map(id => jsonFetch<Volume>(`/api/volumes/${id}`))
  )

  console.log("volumes", volumes);

  for (const { id, name } of volumes) {
    nodes.push({ id, name, type: "volume" });
  }

  console.log("nodes", nodes);
  console.log("links", links);

  forceGraph()
    (document.getElementById('graph'))
    .graphData({ nodes, links })
    .linkDirectionalArrowLength(6)
    .nodeCanvasObject(nodePaint);

}

function nodePaint({ type, x, y, label }: Node, ctx: CanvasRenderingContext2D, globalScale: number) {
  ctx.strokeStyle = "black";
  ctx.lineWidth = 1;
  ctx.fillStyle = "white";
  switch (type) {
    case "network":
      ctx.beginPath();
      ctx.moveTo(x, y - 5);
      ctx.lineTo(x - 5, y + 5);
      ctx.lineTo(x + 5, y + 5);
      ctx.lineTo(x, y - 5);
      ctx.fill();
      ctx.stroke();
      break;
    case "container":
      ctx.beginPath();
      ctx.rect(x - 6, y - 4, 12, 8);
      ctx.fill();
      ctx.stroke();
      break;
    case "volume":
      ctx.beginPath();
      ctx.moveTo(x - 5, y - 5);
      ctx.lineTo(x - 5, y + 5);
      ctx.moveTo(x + 5, y - 5);
      ctx.lineTo(x + 5, y + 5);
      ctx.ellipse(x, y - 5, 5, 2, 0, 0, Math.PI * 2);
      ctx.ellipse(x, y + 5, 5, 2, 0, 0, Math.PI);
      ctx.fill();
      ctx.stroke();
      break;
    default:
      ctx.beginPath();
      ctx.arc(x, y, 5, 0, 2 * Math.PI, false);
      ctx.fill();
      ctx.stroke();
  }
  if (label) {
    // const textWidth = ctx.measureText(label).width;
    // const bckgDimensions = [textWidth, fontSize].map(n => n + fontSize * 0.2); // some padding

    // ctx.fillStyle = 'rgba(255, 255, 255, 0.8)';
    // ctx.fillRect(x - bckgDimensions[0] / 2, y - bckgDimensions[1] / 2, ...bckgDimensions);

    const fontSize = Math.ceil(12 / globalScale);
    ctx.font = `${fontSize}pt Sans-Serif`;
    ctx.textAlign = 'center';
    ctx.textBaseline = 'top';
    ctx.fillStyle = "black";
    ctx.fillText(label, x, y + 5);
  }
}

bootstrap();

console.log("yaay");
