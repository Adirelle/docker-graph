import { LinkCollection } from "../../utils/linkCollection";
import { AnyNode } from "../types";
import { BaseNode } from "./baseNode";

export interface Image {
  ID: string;
  Name: string;
}

export class ImageNode extends BaseNode<Image> {

  public override get icon(): string {
    return "\uf03e";
  }

  public override get type(): string {
    return "image";
  }
}

export class ImageLinks extends LinkCollection<string, Image, ImageNode> {

  protected override mapModel(image: string): Image {
    return { ID: `image:${image}`, Name: image };
  }

  protected override buildNode(model: Image): ImageNode {
    return new ImageNode(model);
  }

  protected override isTargetNode(node: AnyNode): node is ImageNode {
    return node instanceof ImageNode;
  }
}
