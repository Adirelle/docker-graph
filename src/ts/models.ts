import { LinkObject, NodeObject } from "force-graph";

export type NodeType = 'container' | 'network' | 'hostIP' | 'image' | 'bindMount' | 'port' | 'volume';

export interface NodeModel extends NodeObject {
  id: string;
  type: NodeType;
  label: string;
  tooltip?: string;
  color?: string;
  width?: number;
  height?: number;
  links?: NodeModel[];
}

export interface LinkModel extends LinkObject {
  readonly source: NodeModel;
  readonly target: NodeModel;
}

export interface GraphModel {
  getOrCreateNode(id: string): NodeModel;
  removeNode(id: string): void;

  getOrCreateLink(sourceID: string, targetID: string): LinkModel;
  removeLink(link: LinkModel): void;
  listLinksFrom(sourceID: string): Set<LinkModel>;
}
