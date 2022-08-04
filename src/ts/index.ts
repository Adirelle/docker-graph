
import ForceGraph, { NodeObject } from "force-graph";
import type { Event } from "./api";
import { EventProcessor } from "./eventProcessor";
import { GraphData, GraphUpdater } from "./graph";
import { NodeModel } from "./models";
import { NodePainter, TextRenderer } from "./paint/index";
import { consumeEvents, debouncer, Status } from "./utils";

(function (
  graphElem: HTMLElement | null,
  statusElem: HTMLElement | null,
) {
  if (graphElem == null) {
    console.error("could not find graph element");
    return;
  }
  if (statusElem == null) {
    console.error("could not find status element");
    return;
  }

  const graph = new GraphData();
  const processor = new EventProcessor(() => new GraphUpdater(graph));

  const nodePainter = new NodePainter(
    new TextRenderer(),
    new TextRenderer({ font: `FontAwesome`, size: 12 }),
  );

  const forceGraph = ForceGraph();
  forceGraph(graphElem)
    .nodeLabel("tooltip")
    .nodeCanvasObject((node: NodeObject, ctx, scale) => nodePainter.paint(node as NodeModel, ctx, scale))
    .nodePointerAreaPaint((node: NodeObject, color, ctx) => nodePainter.paintInteractionArea(node as NodeModel, color, ctx));

  const trigger = debouncer(300, () => {
    const data = graph.data();
    console.debug("refreshing graph:", data);
    forceGraph.graphData(data);
  });

  consumeEvents(
    "/api/events",
    ({ data }) => {
      const event = JSON.parse(data) as Event;
      console.debug("event", event);
      if (processor.process(event)) {
        trigger();
      }
    },
    (status: Status) => {
      const icon = status == 'open' ? 'wifi' : 'wifi-slash';
      statusElem.className = `fas fa-${icon}`;
    }
  );
})(
  document.getElementById("graph"),
  document.getElementById("status")
);
