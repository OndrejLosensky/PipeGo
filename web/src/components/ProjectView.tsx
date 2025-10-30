import { useQuery } from '@tanstack/react-query';
import { useParams, Link } from 'react-router-dom';
import { api, type PartRunStats } from '../api';

export function ProjectView() {
  const { projectName } = useParams();
  const { data: projects } = useQuery({
    queryKey: ['projects'],
    queryFn: api.getProjects,
  });

  const selectedProject = projectName || projects?.[0]?.name;

  const { data: stats, isLoading } = useQuery({
    queryKey: ['stats', selectedProject],
    queryFn: () => api.getProjectStats(selectedProject!),
    enabled: !!selectedProject,
  });

  const handleTriggerRun = async () => {
    if (!selectedProject) return;
    try {
      await api.triggerRun(selectedProject);
      window.location.reload();
    } catch (error) {
      alert('Failed to trigger run');
    }
  };

  // Group runs by part
  const groupedByPart: Record<string, PartRunStats[]> = {};
  stats?.forEach((stat) => {
    if (!groupedByPart[stat.part]) {
      groupedByPart[stat.part] = [];
    }
    groupedByPart[stat.part].push(stat);
  });

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

        {/* Main content - TeamCity style */}
        <div className="flex-1 p-8">
          {selectedProject && (
            <>
              <div className="flex justify-between items-center mb-6">
                <h2 className="text-2xl font-bold">{selectedProject}</h2>
                <button
                  onClick={handleTriggerRun}
                  className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 text-sm"
                >
                  Run Pipeline
                </button>
              </div>

              {isLoading ? (
                <div className="text-gray-500">Loading...</div>
              ) : (
                <div className="space-y-6">
                  {Object.entries(groupedByPart).map(([part, runs]) => (
                    <div key={part} className="bg-white rounded-lg border border-gray-200">
                      <div className="px-6 py-3 border-b border-gray-200 bg-gray-50">
                        <h3 className="text-lg font-semibold capitalize">{part}</h3>
                      </div>
                      <div className="divide-y divide-gray-100">
                        {runs.map((run) => (
                          <Link
                            key={run.run_id}
                            to={`/projects/${selectedProject}/runs/${run.run_id}`}
                            className="flex items-center justify-between px-6 py-4 hover:bg-gray-50"
                          >
                            <div className="flex items-center gap-4">
                              <span
                                className={`w-3 h-3 rounded-full ${
                                  run.status === 'success'
                                    ? 'bg-green-500'
                                    : run.status === 'failed'
                                    ? 'bg-red-500'
                                    : 'bg-yellow-500'
                                }`}
                              />
                              <div>
                                <div className="font-medium">Run #{run.run_id}</div>
                                <div className="text-sm text-gray-500">
                                  {run.step_count} steps · {formatTime(run.started_at)}
                                </div>
                              </div>
                            </div>
                            <div className="flex items-center gap-6">
                              <div className="text-sm text-gray-600">{run.duration || '-'}</div>
                              <span
                                className={`px-3 py-1 text-xs font-semibold rounded-full ${
                                  run.status === 'success'
                                    ? 'bg-green-100 text-green-800'
                                    : run.status === 'failed'
                                    ? 'bg-red-100 text-red-800'
                                    : 'bg-yellow-100 text-yellow-800'
                                }`}
                              >
                                {run.status}
                              </span>
                            </div>
                          </Link>
                        ))}
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </>
          )}
        </div>
      </div>
    </div>
  );
}
