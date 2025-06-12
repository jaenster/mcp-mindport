export interface Config {
  server: {
    host: string;
    port: number;
  };
  storage: {
    path: string;
  };
  search: {
    indexPath: string;
  };
  domain: {
    defaultDomain: string;
    isolationMode: 'strict' | 'hierarchical';
    allowCrossDomain: boolean;
    currentDomain: string;
  };
}

export interface Resource {
  id: string;
  name: string;
  description?: string;
  content: string;
  tags: string[];
  domain: string;
  mimeType?: string;
  uri?: string;
  createdAt: Date;
  updatedAt: Date;
}

export interface SearchResult {
  resource: Resource;
  score: number;
  matches: string[];
}

export interface Domain {
  name: string;
  description?: string;
  resourceCount: number;
  createdAt: Date;
}

export interface PromptTemplate {
  id: string;
  name: string;
  description?: string;
  template: string;
  variables: string[];
  domain: string;
  createdAt: Date;
  updatedAt: Date;
}