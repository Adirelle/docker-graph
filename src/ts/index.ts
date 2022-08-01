import ForceGraph, { LinkObject, NodeObject } from "force-graph";
import type { Event } from "./api";
import { Hideable, Renderable } from "./graph";
import { GraphData } from "./graph/graphData";

const graphData = new GraphData();

const forceGraph = ForceGraph();
forceGraph(document.getElementById("graph"))
  .nodeCanvasObject((node: NodeObject, ctx, scale) =>
    (node as Renderable).render(ctx, scale)
  )
  .nodePointerAreaPaint(
    (
      node: NodeObject,
      color: string,
      ctx: CanvasRenderingContext2D,
      scale: number
    ) => (node as Renderable).paintInteractionArea(color, ctx, scale)
  )
  .nodeVisibility((node: NodeObject) => (node as Hideable).isVisible())
  .linkVisibility((link: LinkObject) => (link as Hideable).isVisible());

let debounceHandle: any | null;

function update() {
  const data = graphData.data();
  console.debug("refreshing graph:", data);
  forceGraph.graphData(data);
  debounceHandle = null;
}

const es = new EventSource("/api/events");
es.addEventListener("message", ({ data }: { data: string }) => {
  const event = JSON.parse(data) as Event;
  console.debug("event", event);
  if (graphData.process(event)) {
    if (debounceHandle) {
      clearTimeout(debounceHandle);
    }
    debounceHandle = setTimeout(update, 300);
  }
});
