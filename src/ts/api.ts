
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
  Project?: Project;
  Networks?: Networks;
  Mounts?: Mount[];
  Ports?: Ports;
}

export interface Image {
  Registry: string;
  Name: string;
  Tag: string;
}

export interface Project {
  Name: string;
  WorkingDir: string;
}

export interface ContainerList {
  Containers: Array<string>;
}

export interface Networks {
  [path: string]: Network;
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

export interface Ports {
  [def: string]: Port;
}

export interface Port {
  HostIp: string;
  HostPort: number;
}
