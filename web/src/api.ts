const API_BASE = '/api';

export interface Project {
  name: string;
  path: string;
  description?: string;
  valid: boolean;
  error?: string;
}

export interface Run {
  id: number;
  status: string;
  config_path: string;
  project_name: string;
  group: string;
  part: string;
  started_at: string;
  finished_at?: string;
  duration?: string;
}

export interface StepExecution {
  id: number;
  run_id: number;
  name: string;
  status: string;
  command: string;
  output: string;
  group: string;
  part: string;
  category: string;
  started_at: string;
  finished_at?: string;
  duration?: string;
}

export interface RunDetail {
  run: Run;
  steps: StepExecution[];
}

export interface PartRunStats {
  group: string;
  part: string;
  run_id: number;
  status: string;
  duration?: string;
  started_at: string;
  step_count: number;
}

export const api = {
  getProjects: async (): Promise<Project[]> => {
    const res = await fetch(`${API_BASE}/projects`);
    if (!res.ok) throw new Error('Failed to fetch projects');
    return res.json();
  },

  getProjectRuns: async (projectName: string): Promise<Run[]> => {
    const res = await fetch(`${API_BASE}/projects/${projectName}/runs`);
    if (!res.ok) throw new Error('Failed to fetch runs');
    return res.json();
  },

  getProjectStats: async (projectName: string): Promise<PartRunStats[]> => {
    const res = await fetch(`${API_BASE}/projects/${projectName}/stats`);
    if (!res.ok) throw new Error('Failed to fetch project stats');
    return res.json();
  },

  getRun: async (runId: number): Promise<RunDetail> => {
    const res = await fetch(`${API_BASE}/runs/${runId}`);
    if (!res.ok) throw new Error('Failed to fetch run');
    return res.json();
  },

  triggerRun: async (projectName: string, part?: string): Promise<{ run_id: number }> => {
    const url = part 
      ? `${API_BASE}/projects/${projectName}/run?part=${encodeURIComponent(part)}`
      : `${API_BASE}/projects/${projectName}/run`;
    
    const res = await fetch(url, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
    });
    if (!res.ok) throw new Error('Failed to trigger run');
    return res.json();
  },
};

