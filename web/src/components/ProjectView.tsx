import { useQuery, useQueryClient } from '@tanstack/react-query';
import { useParams, Link } from 'react-router-dom';
import { api, type PartRunStats } from '../api';
import { useState, useEffect } from 'react';
import { formatDuration } from '../utils/duration';

export function ProjectView() {
  const { projectName } = useParams();
  const queryClient = useQueryClient();
  const [runningParts, setRunningParts] = useState<Set<string>>(new Set());
  const [triggeringParts, setTriggeringParts] = useState<Set<string>>(new Set());
  
  const { data: projects } = useQuery({
    queryKey: ['projects'],
    queryFn: api.getProjects,
  });

  const selectedProject = projectName || projects?.[0]?.name;

  const { data: stats, isLoading, refetch } = useQuery({
    queryKey: ['stats', selectedProject],
    queryFn: () => api.getProjectStats(selectedProject!),
    enabled: !!selectedProject,
    refetchInterval: runningParts.size > 0 ? 2000 : false,
  });

  // Additional polling check based on actual running status
  const hasRunningRuns = stats?.some(run => run.status === 'running');

  // Set up polling when runs are detected as running
  useEffect(() => {
    if (!hasRunningRuns) return;

    const interval = setInterval(() => {
      refetch();
    }, 2000);

    return () => clearInterval(interval);
  }, [hasRunningRuns, refetch]);

  // Group runs by part (do this early so handlers can use it)
  const groupedByPart: Record<string, PartRunStats[]> = {};
  stats?.forEach((stat) => {
    if (!groupedByPart[stat.part]) {
      groupedByPart[stat.part] = [];
    }
    groupedByPart[stat.part].push(stat);
  });

  // Clear runningParts when no runs are actually running
  useEffect(() => {
    if (!hasRunningRuns && runningParts.size > 0) {
      setRunningParts(new Set());
    }
  }, [hasRunningRuns, runningParts.size]);

  const handleTriggerRun = async (part?: string) => {
    if (!selectedProject) return;
    
    // Prevent double-clicking
    if (part && triggeringParts.has(part)) return;
    
    if (part) {
      setTriggeringParts(prev => new Set(prev).add(part));
      setRunningParts(prev => new Set(prev).add(part));
    }
    
    try {
      await api.triggerRun(selectedProject, part);
      // Small delay to let backend create the run in DB
      await new Promise(resolve => setTimeout(resolve, 300));
      // Force fresh fetch by invalidating cache
      await queryClient.invalidateQueries({ queryKey: ['stats', selectedProject] });
    } catch (error) {
      console.error('Failed to trigger run:', error);
      if (part) {
        setRunningParts(prev => {
          const next = new Set(prev);
          next.delete(part);
          return next;
        });
      }
    } finally {
      if (part) {
        setTriggeringParts(prev => {
          const next = new Set(prev);
          next.delete(part);
          return next;
        });
      }
    }
  };

  const handleTriggerAll = async () => {
    if (!selectedProject || !stats) return;
    
    // Prevent double-clicking
    if (triggeringParts.size > 0) return;

    // Get all unique parts from current stats
    const parts = Array.from(new Set(stats.map(s => s.part)));
    
    // Mark all parts as triggering and running
    setTriggeringParts(new Set(parts));
    setRunningParts(new Set(parts));

    // Trigger all parts in parallel
    try {
      await Promise.all(
        parts.map(part => api.triggerRun(selectedProject, part))
      );
      // Small delay to let backend create the runs in DB
      await new Promise(resolve => setTimeout(resolve, 300));
      // Force fresh fetch by invalidating cache
      await queryClient.invalidateQueries({ queryKey: ['stats', selectedProject] });
    } catch (error) {
      console.error('Failed to trigger runs:', error);
      setRunningParts(new Set());
    } finally {
      setTriggeringParts(new Set());
    }
  };

  // Check if a part has any running runs or is being triggered
  const isPartRunning = (part: string) => {
    return triggeringParts.has(part) ||
           runningParts.has(part) || 
           groupedByPart[part]?.[0]?.status === 'running';
  };

  const formatTime = (dateStr: string) => {
    const date = new Date(dateStr);
    const now = new Date();
    const diff = now.getTime() - date.getTime();
    const minutes = Math.floor(diff / 60000);
    const hours = Math.floor(minutes / 60);
    const days = Math.floor(hours / 24);

    if (minutes < 1) return 'just now';
    if (minutes < 60) return `${minutes}m ago`;
    if (hours < 24) return `${hours}h ago`;
    return `${days}d ago`;
  };

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="flex">
        {/* Sidebar */}
        <div className="w-64 bg-white border-r border-gray-200 min-h-screen p-4">
          <h1 className="text-xl font-bold mb-4">PipeGo</h1>
          <div className="space-y-1">
            {projects?.map((project) => (
              <Link
                key={project.name}
                to={`/projects/${project.name}`}
                className={`block px-3 py-2 rounded text-sm ${
                  project.name === selectedProject
                    ? 'bg-blue-50 text-blue-700 font-medium'
                    : 'text-gray-700 hover:bg-gray-50'
                }`}
              >
                {project.name}
              </Link>
            ))}
          </div>
        </div>

        {/* Main content */}
        <div className="flex-1 p-8">
          {selectedProject && (
            <>
              <div className="flex justify-between items-center mb-6 pb-4 border-b border-gray-200">
                <h2 className="text-xl font-bold text-gray-900">{selectedProject}</h2>
                <button
                  onClick={handleTriggerAll}
                  disabled={triggeringParts.size > 0 || runningParts.size > 0}
                  className={`px-4 py-1.5 rounded text-sm font-medium flex items-center gap-2 transition-colors ${
                    triggeringParts.size > 0 || runningParts.size > 0
                      ? 'bg-gray-300 text-gray-600 cursor-not-allowed'
                      : 'bg-blue-600 text-white hover:bg-blue-700'
                  }`}
                >
                  {(triggeringParts.size > 0 || runningParts.size > 0) && (
                    <div className="w-3.5 h-3.5 border-2 border-white border-t-transparent rounded-full animate-spin" />
                  )}
                  {triggeringParts.size > 0 || runningParts.size > 0 ? 'Running...' : 'Run All'}
                </button>
              </div>

              {isLoading && !stats ? (
                <div className="text-gray-500">Loading...</div>
              ) : (
                <div className="space-y-6">
                  {Object.entries(groupedByPart).map(([part, runs]) => {
                    const partRunning = isPartRunning(part);
                    
                    return (
                      <div key={part} className="bg-white rounded-lg border border-gray-200 shadow-sm">
                        <div className="px-6 py-2.5 border-b border-gray-200 flex items-center justify-between">
                          <h3 className="text-base font-semibold text-gray-900 capitalize flex items-center gap-2">
                            {part}
                            {partRunning && (
                              <div className="w-3.5 h-3.5 border-2 border-blue-600 border-t-transparent rounded-full animate-spin" />
                            )}
                          </h3>
                          <button
                            onClick={() => handleTriggerRun(part)}
                            disabled={partRunning}
                            className={`px-3 py-1 text-xs font-medium rounded transition-colors ${
                              partRunning
                                ? 'bg-gray-200 text-gray-500 cursor-not-allowed'
                                : 'bg-green-600 text-white hover:bg-green-700'
                            }`}
                          >
                            {partRunning ? 'Running...' : 'Run'}
                          </button>
                        </div>
                        <div className="divide-y divide-gray-100">
                          {runs.map((run) => (
                            <Link
                              key={run.run_id}
                              to={`/projects/${selectedProject}/runs/${run.run_id}`}
                              className="flex items-center justify-between px-6 py-2.5 hover:bg-gray-50 transition-colors"
                            >
                              <div className="flex items-center gap-3">
                                {/* Status indicator dot */}
                                <div className="relative flex items-center justify-center w-2 h-2">
                                  <span
                                    className={`block w-2 h-2 rounded-full ${
                                      run.status === 'success'
                                        ? 'bg-green-500'
                                        : run.status === 'failed'
                                        ? 'bg-red-500'
                                        : 'bg-yellow-500'
                                    }`}
                                  />
                                  {run.status === 'running' && (
                                    <span className="absolute w-2 h-2 rounded-full bg-yellow-400 animate-ping" />
                                  )}
                                </div>
                                
                                {/* Run info */}
                                <div className="flex items-center gap-2">
                                  <span className="font-medium text-gray-900">Run #{run.run_id}</span>
                                  <span
                                    className={`px-2 py-0.5 text-xs font-medium rounded ${
                                      run.status === 'success'
                                        ? 'bg-green-100 text-green-700'
                                        : run.status === 'failed'
                                        ? 'bg-red-100 text-red-700'
                                        : 'bg-yellow-100 text-yellow-700'
                                    }`}
                                  >
                                    {run.status}
                                  </span>
                                </div>
                              </div>

                              {/* Right side: steps, time, duration */}
                              <div className="flex items-center gap-3 text-sm text-gray-600 font-mono">
                                <span className="w-16 text-right">{run.step_count} steps</span>
                                <span className="text-gray-400">·</span>
                                <span className="w-20 text-right">{formatTime(run.started_at)}</span>
                                <span className="text-gray-400">·</span>
                                <span className="w-16 text-right font-medium">
                                  {formatDuration(run.duration)}
                                </span>
                              </div>
                            </Link>
                          ))}
                        </div>
                      </div>
                    );
                  })}
                </div>
              )}
            </>
          )}
        </div>
      </div>
    </div>
  );
}
