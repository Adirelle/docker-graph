import ForceGraph, { NodeObject } from "force-graph";
import type { Event } from "./api";
import { EventProcessor } from "./eventProcessor";
import { GraphData, Mapper, Updater } from "./graph";
import { NodeModel } from "./models";
import { NodePainter, TextRenderer } from "./paint/index";

const graph = new GraphData();
const updater = new Updater(graph);
const mapper = new Mapper();
const processor = new EventProcessor(graph, mapper, updater);

const nodePainter = new NodePainter(
  new TextRenderer(),
  new TextRenderer({ font: `FontAwesome`, size: 12 }),
);

const forceGraph = ForceGraph();
forceGraph(document.getElementById("graph"))
  .nodeLabel("tooltip")
  .nodeCanvasObject((node: NodeObject, ctx, scale) => nodePainter.paint(node as NodeModel, ctx, scale))
  .nodePointerAreaPaint((node: NodeObject, color, ctx) => nodePainter.paintInteractionArea(node as NodeModel, color, ctx));

let debounceHandle: any | null;

function update() {
  const data = graph.data();
  console.debug("refreshing graph:", data);
  forceGraph.graphData(data);
  debounceHandle = null;
}

const es = new EventSource("/api/events");
es.addEventListener("message", ({ data }: { data: string; }) => {
  const event = JSON.parse(data) as Event;
  console.debug("event", event);
  if (processor.process(event)) {
    if (debounceHandle) {
      clearTimeout(debounceHandle);
    }
    debounceHandle = setTimeout(update, 300);
  }
});
