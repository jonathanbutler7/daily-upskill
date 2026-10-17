// Our React components:

// const ComponentA = () => {
//   const [count, setCount] = useAtom<number>(counterAtom);

//   return <button onClick={() => setCount(count + 1)}>A: {count}</button>;
// };

// const ComponentB = () => {
//   const [count] = useAtom<number>(counterAtom);

//   return <div>B: {count}</div>;
// };

//Together, we will come up with an implementation for the Atom class and the useAtom hook:
// function useAtom<T>(atom: Atom<T>): [T | null, (val: T) => void] {
//   const [value, setValue] = useState(atom.get());
//   useEffect(() => {
//     const id = atom.subscribe((val) => setValue(val));
//     return () => {
//       atom.unsubscribe(id);
//     };
//   }, [atom]);
//   return [value, (newValue: T) => atom.set(newValue)];
// }

type Subscriber<T> = {
  id: number;
  fn: (val: T) => void;
};
class Atom<T> {
  val: T | null = null;
  subscribers: Subscriber<T>[] = [];
  id = 0;
  set(val: T) {
    this.val = val;
    this.subscribers.forEach((subscriber) => {
      subscriber.fn(val);
    });
  }
  get() {
    return this.val;
  }
  subscribe(cb: Subscriber<T>['fn']) {
    const currentId = this.id;
    this.subscribers = [...this.subscribers, { id: currentId, fn: cb }];
    this.id++;
    return currentId;
  }
  unsubscribe(id: number) {
    this.subscribers = this.subscribers.filter((s) => s.id !== id);
  }
}
// Similarly, the atom should be consumable by any Javascript context in which we wish to use it:

const counterAtom = new Atom<number>();

// Subscribe to changes on the atom.
const id = counterAtom.subscribe((newValue) => {
  // Do something with the new value.
  console.log('oh hey value was set', newValue);
});

// Set the value of the atom.
counterAtom.set(5);

// Get the value of the atom.
const i = counterAtom.get();
console.log(i);
// Unsubscribe.
counterAtom.unsubscribe(id);