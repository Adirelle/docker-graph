import { TranscodeEncoding } from "buffer";

export interface EventBase {
  TargetID: string;
  TargetType: string;
  Type: string;
  Time: string;
  Details?: object;
}

export interface UpdatedEvent extends EventBase {
  Type: "updated";
}

export interface RemovedEvent extends EventBase {
  Type: "removed";
}

export interface ContainerUpdated extends UpdatedEvent {
  TargetType: "container";
  Details: Container;
}

export type Event = ContainerUpdated | RemovedEvent;

export interface Container {
  ID: string;
  Name: string;
  Status: string;
  Image: string;
  Healty: string;
  Service?: string;
  Project?: string;
  Networks?: { [path: string]: Network };
  Mounts?: Mount[];
  Ports?: { [def: string]: Port };
}

export interface ContainerList {
  Containers: Array<string>;
}

export interface Network {
  ID: string;
  Name: string;
}

export interface Mount {
  Name: string;
  Type: string;
  Source: string;
  Destination: string;
  ReadWrite: boolean;
}

export interface Volume {
  Id: string;
  Name: string;
}

export interface Port {
  HostIp: string;
  HostPort: number;
}
