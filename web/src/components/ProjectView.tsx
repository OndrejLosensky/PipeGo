import { useQuery } from '@tanstack/react-query';
import { useParams, Link } from 'react-router-dom';
import { api } from '../api';

export function ProjectView() {
  const { projectName } = useParams();
  const { data: projects } = useQuery({
    queryKey: ['projects'],
    queryFn: api.getProjects,
  });

  const selectedProject = projectName || projects?.[0]?.name;

  const { data: runs, isLoading } = useQuery({
    queryKey: ['runs', selectedProject],
    queryFn: () => api.getProjectRuns(selectedProject!),
    enabled: !!selectedProject,
  });

  const handleTriggerRun = async () => {
    if (!selectedProject) return;
    try {
      await api.triggerRun(selectedProject);
      window.location.reload();
    } catch (error) {
      console.error(`Failed to trigger run: ${error}`);
    }
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
                <div className="bg-white rounded-lg border border-gray-200">
                  <table className="w-full">
                    <thead className="bg-gray-50 border-b border-gray-200">
                      <tr>
                        <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                          ID
                        </th>
                        <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                          Part
                        </th>
                        <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                          Status
                        </th>
                        <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                          Started
                        </th>
                        <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                          Duration
                        </th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-gray-200">
                      {runs?.map((run) => (
                        <tr key={run.id} className="hover:bg-gray-50">
                          <td className="px-4 py-3">
                            <Link
                              to={`/projects/${selectedProject}/runs/${run.id}`}
                              className="text-blue-600 hover:underline"
                            >
                              #{run.id}
                            </Link>
                          </td>
                          <td className="px-4 py-3 text-sm text-gray-600">
                            {run.part === 'default' ? '-' : run.part}
                          </td>
                          <td className="px-4 py-3">
                            <span
                              className={`inline-flex px-2 py-1 text-xs font-semibold rounded ${
                                run.status === 'success'
                                  ? 'bg-green-100 text-green-800'
                                  : run.status === 'failed'
                                  ? 'bg-red-100 text-red-800'
                                  : 'bg-yellow-100 text-yellow-800'
                              }`}
                            >
                              {run.status}
                            </span>
                          </td>
                          <td className="px-4 py-3 text-sm text-gray-600">
                            {new Date(run.started_at).toLocaleString()}
                          </td>
                          <td className="px-4 py-3 text-sm text-gray-600">
                            {run.duration || '-'}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </>
          )}
        </div>
      </div>
    </div>
  );
}

