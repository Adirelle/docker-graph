import { Project } from "./api";

export function* iterMap<T, U>(inputs: Iterable<T>, f: (input: T) => U): Iterable<U> {
  for (const input of inputs) {
    yield f(input);
  }
}

export function* iterFlatMap<T, U>(inputs: Iterable<T>, f: (input: T) => Iterable<U>): Iterable<U> {
  for (const input of inputs) {
    for (const output of f(input)) {
      yield output;
    }
  }
}


const idRegexp = /^[0-9a-f]{64}$/;

export function isID(what: string): boolean {
  return idRegexp.test(what);
}

export function shortID(id: string): string {
  return id.substring(0, 8);
}

export function shortIDOrName(idOrName: string): string {
  if (isID(idOrName)) {
    return shortID(idOrName);
  }
  return idOrName;
}

export function shortName(idOrName: string, project: Project | undefined): string {
  if (isID(idOrName)) {
    return shortID(idOrName);
  }
  if (project && (idOrName.startsWith(project.Name + "_") || idOrName.startsWith(project.Name + "-"))) {
    return idOrName.substring(project.Name.length + 1);
  }
  return idOrName;
}

export function shortPath(path: string, project: Project | undefined): string {
  if (project) {
    console.log("shortPath", path, project.WorkingDir);
    if (path == project.WorkingDir) {
      return ".";
    }
    if (path.startsWith(project.WorkingDir + "/")) {
      return "." + path.substring(project.WorkingDir.length);
    }
  }
  return path;
}
