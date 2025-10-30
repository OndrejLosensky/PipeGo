import { useQuery } from '@tanstack/react-query';
import { useParams, Link } from 'react-router-dom';
import { api } from '../api';
import { useState, useEffect } from 'react';
import { formatDuration } from '../utils/duration';

export function RunDetail() {
  const { projectName, runId } = useParams();
  const [expandedStep, setExpandedStep] = useState<number | null>(null);

  // Real-time polling - refetch every 2s when run is active
  const { data, isLoading } = useQuery({
    queryKey: ['run', runId],
    queryFn: () => api.getRun(Number(runId)),
    enabled: !!runId,
    refetchInterval: (query) => {
      return query.state.data?.run.status === 'running' ? 2000 : false;
    },
  });

  // Auto-expand failed steps
  useEffect(() => {
    if (data?.steps) {
      const failedStep = data.steps.find(s => s.status === 'failed');
      if (failedStep && expandedStep === null) {
        setExpandedStep(failedStep.id);
      }
    }
  }, [data, expandedStep]);

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

        {/* Header - more compact */}
        <div className="bg-white rounded-lg border border-gray-200 p-5 mb-6 shadow-sm">
          <h1 className="text-xl font-bold mb-3">Run #{data.run.id}</h1>
          <div className="grid grid-cols-2 gap-3 text-sm">
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
                className={`ml-2 inline-flex px-2 py-0.5 text-xs font-medium rounded ${
                  data.run.status === 'success'
                    ? 'bg-green-100 text-green-700'
                    : data.run.status === 'failed'
                    ? 'bg-red-100 text-red-700'
                    : 'bg-yellow-100 text-yellow-700'
                }`}
              >
                {data.run.status}
              </span>
            </div>
            <div>
              <span className="text-gray-500">Duration:</span>
              <span className="ml-2 font-medium">{formatDuration(data.run.duration)}</span>
            </div>
          </div>
        </div>

        <h2 className="text-lg font-bold mb-3">Steps</h2>
        <div className="space-y-2">
          {data.steps.map((step) => {
            const isRunning = step.status === 'running';
            const isFailed = step.status === 'failed';
            const isSuccess = step.status === 'success';
            
            return (
              <div
                key={step.id}
                className={`bg-white rounded-lg border overflow-hidden transition-colors ${
                  isRunning 
                    ? 'border-blue-300' 
                    : isFailed 
                    ? 'border-red-200' 
                    : 'border-gray-200'
                }`}
              >
                <div
                  className="flex items-center justify-between py-2.5 px-4 cursor-pointer hover:bg-gray-50 transition-colors"
                  onClick={() =>
                    setExpandedStep(expandedStep === step.id ? null : step.id)
                  }
                >
                  <div className="flex items-center gap-3">
                    {/* Status indicator */}
                    <div className="flex items-center justify-center w-5 h-5">
                      {isRunning && (
                        <div className="w-4 h-4 border-2 border-blue-600 border-t-transparent rounded-full animate-spin" />
                      )}
                      {isSuccess && (
                        <svg className="w-5 h-5 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                        </svg>
                      )}
                      {isFailed && (
                        <svg className="w-5 h-5 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                        </svg>
                      )}
                    </div>
                    
                    <div>
                      <div className="font-medium text-gray-900">{step.name}</div>
                      {step.category && (
                        <div className="text-xs text-gray-500">{step.category}</div>
                      )}
                    </div>
                  </div>
                  <div className="text-sm text-gray-600 font-mono">
                    {formatDuration(step.duration)}
                  </div>
                </div>
                
                {/* Log output - enhanced styling */}
                {expandedStep === step.id && step.output && (
                  <div className="border-t border-gray-200 bg-gray-950 p-3">
                    <pre className="text-green-400 text-xs font-mono overflow-x-auto whitespace-pre-wrap leading-relaxed">
                      {step.output}
                    </pre>
                  </div>
                )}
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}
