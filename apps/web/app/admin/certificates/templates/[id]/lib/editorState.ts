import type {
  CertificateDocument,
  CertificateElement,
  CertificateTemplate,
} from "../../../lib/template/model";

const HISTORY_LIMIT = 100;
const COALESCE_MS = 700;

export type EditorState = {
  template: CertificateTemplate;
  document: CertificateDocument;
  name: string;
  description: string;
  selectedId: string | null;
  past: CertificateDocument[];
  future: CertificateDocument[];
  dirty: boolean;
  lastKey: string | null;
  lastAt: number;
};

export type EditorAction =
  | { type: "change"; document: CertificateDocument; key?: string; at: number }
  | { type: "preview"; document: CertificateDocument }
  | { type: "commit"; before: CertificateDocument }
  | { type: "undo" }
  | { type: "redo" }
  | { type: "select"; id: string | null }
  | { type: "rename"; name: string }
  | { type: "describe"; description: string }
  | {
      type: "saved";
      template: CertificateTemplate;
      document: CertificateDocument;
      name: string;
      description: string;
    }
  | { type: "replace"; template: CertificateTemplate };

export function initialState(template: CertificateTemplate): EditorState {
  return {
    template,
    document: template.document,
    name: template.name,
    description: template.description,
    selectedId: null,
    past: [],
    future: [],
    dirty: false,
    lastKey: null,
    lastAt: 0,
  };
}

function pushed(past: CertificateDocument[], document: CertificateDocument) {
  return [...past, document].slice(-HISTORY_LIMIT);
}

export function editorReducer(state: EditorState, action: EditorAction): EditorState {
  switch (action.type) {
    case "change": {
      const coalesce =
        action.key !== undefined && action.key === state.lastKey && action.at - state.lastAt < COALESCE_MS;
      return {
        ...state,
        document: action.document,
        past: coalesce ? state.past : pushed(state.past, state.document),
        future: [],
        dirty: true,
        lastKey: action.key ?? null,
        lastAt: action.at,
      };
    }
    case "preview":
      return { ...state, document: action.document };
    case "commit":
      if (action.before === state.document) return state;
      return {
        ...state,
        past: pushed(state.past, action.before),
        future: [],
        dirty: true,
        lastKey: null,
      };
    case "undo": {
      const previous = state.past.at(-1);
      if (!previous) return state;
      return {
        ...state,
        document: previous,
        past: state.past.slice(0, -1),
        future: [state.document, ...state.future],
        dirty: true,
        lastKey: null,
        selectedId: previous.elements.some((e) => e.id === state.selectedId) ? state.selectedId : null,
      };
    }
    case "redo": {
      const next = state.future[0];
      if (!next) return state;
      return {
        ...state,
        document: next,
        past: pushed(state.past, state.document),
        future: state.future.slice(1),
        dirty: true,
        lastKey: null,
        selectedId: next.elements.some((e) => e.id === state.selectedId) ? state.selectedId : null,
      };
    }
    case "select":
      return { ...state, selectedId: action.id, lastKey: null };
    case "rename":
      return { ...state, name: action.name, dirty: true };
    case "describe":
      return { ...state, description: action.description, dirty: true };
    case "saved":
      return {
        ...state,
        template: action.template,
        dirty:
          state.document !== action.document ||
          state.name !== action.name ||
          state.description !== action.description,
      };
    case "replace":
      return initialState(action.template);
  }
}

export function updateElement(
  document: CertificateDocument,
  id: string,
  update: (element: CertificateElement) => CertificateElement,
): CertificateDocument {
  return {
    ...document,
    elements: document.elements.map((element) => (element.id === id ? update(element) : element)),
  };
}

export function removeElement(document: CertificateDocument, id: string): CertificateDocument {
  return { ...document, elements: document.elements.filter((element) => element.id !== id) };
}

export function insertElement(
  document: CertificateDocument,
  element: CertificateElement,
  index = document.elements.length,
): CertificateDocument {
  const elements = [...document.elements];
  elements.splice(index, 0, element);
  return { ...document, elements };
}

export type LayerMove = "front" | "back" | "forward" | "backward";

export function moveLayer(document: CertificateDocument, id: string, move: LayerMove): CertificateDocument {
  const index = document.elements.findIndex((element) => element.id === id);
  if (index < 0) return document;
  const elements = [...document.elements];
  const [element] = elements.splice(index, 1);
  const target =
    move === "front"
      ? elements.length
      : move === "back"
        ? 0
        : move === "forward"
          ? Math.min(index + 1, elements.length)
          : Math.max(index - 1, 0);
  elements.splice(target, 0, element);
  return { ...document, elements };
}
