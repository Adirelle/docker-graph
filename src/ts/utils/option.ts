
export abstract class Option<T> implements Iterable<T> {
  public abstract isSome(): boolean;
  public abstract isNone(): boolean;
  public abstract filter(predicate: (x: T) => Boolean): Option<T>;
  public abstract map<U>(f: (x: T) => U): Option<U>;
  public abstract flatMap<U>(f: (x: T) => Option<U>): Option<U>;
  public abstract getOr(defaultValue: T): T;
  public abstract getOrThrow(): T;
  public abstract [Symbol.iterator](): Iterator<T, any, undefined>;

  public static From<T>(value: T | null | undefined): Option<T> {
    if (value === undefined || value === null) {
      return None();
    }
    return Some(value);
  }

  public static First<T>(...items: T[]): Option<T>;
  public static First<T>(items: Iterable<T>): Option<T> {
    for (const value of items) {
      return Some(value);
    }
    return None();
  }
}

class some<T> extends Option<T> {
  public constructor(private readonly value: T) { super(); }
  public override isSome(): boolean { return true; }
  public override isNone(): boolean { return false; }
  public override filter(predicate: (x: T) => Boolean): Option<T> { return predicate(this.value) ? this : None(); };
  public override map<U>(f: (x: T) => U): Option<U> { return new some(f(this.value)); }
  public override flatMap<U>(f: (x: T) => Option<U>): Option<U> { return f(this.value); }
  public override getOr(_defaultValue: T): T { return this.value; }
  public override getOrThrow(): T { return this.value; }
  public override *[Symbol.iterator](): Iterator<T, any, undefined> {
    yield this.value;
  }
}

export function Some<T>(value: T): Option<T> {
  return new some(value);
}

class none<T> extends Option<T> {
  public override filter(_predicate: (x: T) => Boolean): Option<T> { return this; }
  public override isSome(): boolean { return false; }
  public override isNone(): boolean { return true; }
  public override map<U>(_f: (x: any) => U): Option<U> { return None(); };
  public override flatMap<U>(_f: (x: any) => Option<U>): Option<U> { return None(); };
  public override getOr(defaultValue: T) { return defaultValue; };
  public override getOrThrow(): never { throw new NoneError(); }
  public override *[Symbol.iterator](): Iterator<T, any, undefined> { }
}

const theNone = new none();

export class NoneError extends Error {
  public constructor() { super("No value"); }
}

export function None<T>(): Option<T> {
  return theNone as Option<T>;
}
