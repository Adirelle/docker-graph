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

export function debouncer(delay: number, proc: () => void): () => void {
  let handle: number | null;
  let callback = () => {
    handle = null;
    proc();
  };
  return () => {
    if (handle) {
      clearTimeout(handle);
    }
    handle = setTimeout(callback, delay);
  };
}

export type Status = 'open' | 'closed';

export function consumeEvents(sourceURL: string, handler: (ev: MessageEvent) => void, statusHandler: (st: Status) => void = () => null): void {
  let restartHandle: number | null = null;
  const run = () => {
    const source = new EventSource(sourceURL);
    source.addEventListener("message", handler);
    source.addEventListener("open", () => {
      statusHandler('open');
      restartHandle = null;
    });
    source.addEventListener("error", () => {
      source.close();
      statusHandler('closed');
      if (!restartHandle) {
        restartHandle = setTimeout(run, 10);
      }
    });
  };
  run();
}
