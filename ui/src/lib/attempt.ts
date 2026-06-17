import { fromPromise } from "neverthrow"

export function attempt<T, E = Error>(promise: Promise<T>) {
   return fromPromise(promise, err => err as E)
}
