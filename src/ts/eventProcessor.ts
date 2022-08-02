import { Event } from "./api";
import { Mapper } from "./graph/mapper";
import { Updater } from "./graph/updater";
import { GraphModel } from "./models";

export class EventProcessor {

  public constructor(
    private readonly graph: GraphModel,
    private readonly mapper: Mapper,
    private readonly updater: Updater
  ) { }

  public process(event: Event): boolean {
    if (event.TargetType != "container") {
      return false;
    }
    if (event.Type == "removed") {
      this.graph.removeNode(event.TargetID);
    } else {
      this.updater.update(this.mapper.map(event.Details));
    }
    return true;
  }
}
