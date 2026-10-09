export type LearnNode = {
  id: string;
  parentId: string | null;
  slug: string;
  title: string;
  description: string;
  icon: string;
  imageUrl: string;
  path: string[];
  childCount: number;
  algorithmCount: number;
};

export type LearnAlgorithm = {
  id: string;
  nodeId: string;
  notation: string;
  name: string;
  notes: string;
};

export type LearnLevel = {
  node: LearnNode | null;
  ancestors: LearnNode[];
  children: LearnNode[];
  algorithms: LearnAlgorithm[];
};
