import { useQuery } from '@tanstack/react-query';
import { useParams, Link } from 'react-router-dom';
import { api } from '../api';
import { useState } from 'react';

export function RunDetail() {
  const { projectName, runId } = useParams();
  const [expandedStep, setExpandedStep] = useState<number | null>(null);

  const { data, isLoading } = useQuery({
    queryKey: ['run', runId],
    queryFn: () => api.getRun(Number(runId)),
    enabled: !!runId,
  });

  if (isLoading) return <div className="p-8">Loading...</div>;
  if (!data) return <div className="p-8">Run not found</div>;

  return (
    <div className="min-h-screen bg-gray-50 p-8">
      <div className="max-w-6xl mx-auto">
        <Link
          to={`/projects/${projectName}`}
          className="text-blue-600 hover:underline mb-4 inline-block"
        >
          ← Back to {projectName}
        </Link>

        <div className="bg-white rounded-lg border border-gray-200 p-6 mb-6">
          <h1 className="text-2xl font-bold mb-4">Run #{data.run.id}</h1>
          <div className="grid grid-cols-2 gap-4 text-sm">
            <div>
              <span className="text-gray-500">Project:</span>
              <span className="ml-2 font-medium">{data.run.project_name}</span>
            </div>
            <div>
              <span className="text-gray-500">Part:</span>
              <span className="ml-2 font-medium">
                {data.run.part === 'default' ? '-' : data.run.part}
              </span>
            </div>
            <div>
              <span className="text-gray-500">Status:</span>
              <span
                className={`ml-2 inline-flex px-2 py-1 text-xs font-semibold rounded ${
                  data.run.status === 'success'
                    ? 'bg-green-100 text-green-800'
                    : data.run.status === 'failed'
                    ? 'bg-red-100 text-red-800'
                    : 'bg-yellow-100 text-yellow-800'
                }`}
              >
                {data.run.status}
              </span>
            </div>
            <div>
              <span className="text-gray-500">Duration:</span>
              <span className="ml-2 font-medium">{data.run.duration || '-'}</span>
            </div>
          </div>
        </div>

        <h2 className="text-xl font-bold mb-4">Steps</h2>
        <div className="space-y-2">
          {data.steps.map((step) => (
            <div
              key={step.id}
              className="bg-white rounded-lg border border-gray-200 overflow-hidden"
            >
              <div
                className="flex items-center justify-between p-4 cursor-pointer hover:bg-gray-50"
                onClick={() =>
                  setExpandedStep(expandedStep === step.id ? null : step.id)
                }
              >
                <div className="flex items-center gap-3">
                  <span
                    className={`text-lg ${
                      step.status === 'success'
                        ? 'text-green-600'
                        : step.status === 'failed'
                        ? 'text-red-600'
                        : 'text-yellow-600'
                    }`}
                  >
                    {step.status === 'success'
                      ? '✓'
                      : step.status === 'failed'
                      ? '✗'
                      : '⏳'}
                  </span>
                  <div>
                    <div className="font-medium">{step.name}</div>
                    {step.category && (
                      <div className="text-xs text-gray-500">{step.category}</div>
                    )}
                  </div>
                </div>
                <div className="text-sm text-gray-600">{step.duration || '-'}</div>
              </div>
              {expandedStep === step.id && step.output && (
                <div className="border-t border-gray-200 p-4 bg-black text-green-400 font-mono text-xs overflow-x-auto">
                  <pre className="whitespace-pre-wrap">{step.output}</pre>
                </div>
              )}
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

