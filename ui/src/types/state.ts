export type StateTuple<T> = readonly [T, React.Dispatch<React.SetStateAction<T>>]

export type SetterActions<T> = {
   [K in keyof T as `set${Capitalize<string & K>}`]: (value: T[K]) => void
}
