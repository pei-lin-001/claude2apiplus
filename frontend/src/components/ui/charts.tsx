import React, { useRef, useEffect } from 'react';
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  BarElement,
  Title,
  Tooltip,
  Legend,
  ArcElement,
} from 'chart.js';
import { Line, Bar, Doughnut } from 'react-chartjs-2';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { AlertTriangle, CheckCircle, Zap, Shield, Timer, Activity } from 'lucide-react';

// 注册 Chart.js 组件
ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  BarElement,
  Title,
  Tooltip,
  Legend,
  ArcElement
);

interface ChartCardProps {
  title: string;
  description?: string;
  children: React.ReactNode;
}

const ChartCard: React.FC<ChartCardProps> = ({ title, description, children }) => (
  <Card>
    <CardHeader>
      <CardTitle>{title}</CardTitle>
      {description && <CardDescription>{description}</CardDescription>}
    </CardHeader>
    <CardContent>{children}</CardContent>
  </Card>
);

interface SessionStatusChartProps {
  sessions: Array<{
    status: number; // 0: active, 1: cooling, 2: failed, 3: circuit_open
  }>;
}

export const SessionStatusChart: React.FC<SessionStatusChartProps> = ({ sessions }) => {
  const statusCounts = sessions.reduce((acc, session) => {
    acc[session.status] = (acc[session.status] || 0) + 1;
    return acc;
  }, {} as Record<number, number>);

  const data = {
    labels: ['活跃', '冷却', '失败', '熔断'],
    datasets: [
      {
        data: [
          statusCounts[0] || 0, // active
          statusCounts[1] || 0, // cooling
          statusCounts[2] || 0, // failed
          statusCounts[3] || 0, // circuit_open
        ],
        backgroundColor: [
          'rgba(34, 197, 94, 0.8)', // green
          'rgba(251, 191, 36, 0.8)', // yellow
          'rgba(239, 68, 68, 0.8)', // red
          'rgba(239, 68, 68, 0.6)', // dark red
        ],
        borderColor: [
          'rgba(34, 197, 94, 1)',
          'rgba(251, 191, 36, 1)',
          'rgba(239, 68, 68, 1)',
          'rgba(239, 68, 68, 1)',
        ],
        borderWidth: 1,
      },
    ],
  };

  const options = {
    responsive: true,
    plugins: {
      legend: {
        position: 'bottom' as const,
      },
    },
  };

  return (
    <ChartCard title="Session状态分布" description="当前所有 Session 的状态统计">
      <div className="h-64 flex items-center justify-center">
        <Doughnut data={data} options={options} />
      </div>
    </ChartCard>
  );
};

interface ResponseTimeChartProps {
  sessions: Array<{
    avg_response_time: number;
    session_key: string;
  }>;
}

export const ResponseTimeChart: React.FC<ResponseTimeChartProps> = ({ sessions }) => {
  const sortedSessions = [...sessions]
    .sort((a, b) => b.avg_response_time - a.avg_response_time)
    .slice(0, 10); // 只显示前 10 个

  const data = {
    labels: sortedSessions.map((session) => session.session_key.substring(0, 8) + '...'),
    datasets: [
      {
        label: '平均响应时间 (ms)',
        data: sortedSessions.map((session) => session.avg_response_time),
        backgroundColor: 'rgba(59, 130, 246, 0.5)',
        borderColor: 'rgba(59, 130, 246, 1)',
        borderWidth: 1,
      },
    ],
  };

  const options = {
    responsive: true,
    plugins: {
      legend: {
        display: false,
      },
    },
    scales: {
      y: {
        beginAtZero: true,
      },
    },
  };

  return (
    <ChartCard title="响应时间排行" description="Session 平均响应时间对比">
      <div className="h-64">
        <Bar data={data} options={options} />
      </div>
    </ChartCard>
  );
};

interface SuccessRateChartProps {
  sessions: Array<{
    success_count: number;
    total_requests: number;
    session_key: string;
  }>;
}

export const SuccessRateChart: React.FC<SuccessRateChartProps> = ({ sessions }) => {
  const successRates = sessions
    .map((session) => ({
      session: session.session_key.substring(0, 8) + '...',
      rate: session.total_requests > 0 ? (session.success_count / session.total_requests) * 100 : 0,
    }))
    .sort((a, b) => b.rate - a.rate);

  const data = {
    labels: successRates.map((item) => item.session),
    datasets: [
      {
        label: '成功率 (%)',
        data: successRates.map((item) => item.rate),
        borderColor: 'rgba(34, 197, 94, 1)',
        backgroundColor: 'rgba(34, 197, 94, 0.1)',
        tension: 0.1,
        fill: true,
      },
    ],
  };

  const options = {
    responsive: true,
    scales: {
      y: {
        beginAtZero: true,
        max: 100,
        title: {
          display: true,
          text: '成功率 (%)',
        },
      },
    },
    plugins: {
      legend: {
        display: false,
      },
    },
  };

  return (
    <ChartCard title="成功率趋势" description="各 Session 的成功率对比">
      <div className="h-64">
        <Line data={data} options={options} />
      </div>
    </ChartCard>
  );
};

interface HealthScoreChartProps {
  sessions: Array<{
    health_score: number;
    session_key: string;
  }>;
}

export const HealthScoreChart: React.FC<HealthScoreChartProps> = ({ sessions }) => {
  const healthScores = sessions
    .map((session) => ({
      session: session.session_key.substring(0, 8) + '...',
      score: session.health_score * 100, // 转换为百分比
    }))
    .sort((a, b) => b.score - a.score);

  const data = {
    labels: healthScores.map((item) => item.session),
    datasets: [
      {
        label: '健康分数',
        data: healthScores.map((item) => item.score),
        backgroundColor: healthScores.map((item) => {
          if (item.score >= 80) return 'rgba(34, 197, 94, 0.8)';
          if (item.score >= 60) return 'rgba(251, 191, 36, 0.8)';
          return 'rgba(239, 68, 68, 0.8)';
        }),
        borderColor: healthScores.map((item) => {
          if (item.score >= 80) return 'rgba(34, 197, 94, 1)';
          if (item.score >= 60) return 'rgba(251, 191, 36, 1)';
          return 'rgba(239, 68, 68, 1)';
        }),
        borderWidth: 1,
      },
    ],
  };

  const options = {
    responsive: true,
    scales: {
      y: {
        beginAtZero: true,
        max: 100,
        title: {
          display: true,
          text: '健康分数',
        },
      },
    },
    plugins: {
      legend: {
        display: false,
      },
    },
  };

  return (
    <ChartCard title="健康分数排行" description="Session 健康分数对比">
      <div className="h-64">
        <Bar data={data} options={options} />
      </div>
    </ChartCard>
  );
};

interface CallCountChartProps {
  call_count_by_hour: Record<string, number>;
}

export const CallCountChart: React.FC<CallCountChartProps> = ({ call_count_by_hour }) => {
  // 获取最近24小时的数据
  const hours = [];
  const counts = [];
  
  for (let i = 23; i >= 0; i--) {
    const hour = new Date();
    hour.setHours(hour.getHours() - i);
    const hourKey = hour.toISOString().slice(0, 13).replace('T', ' ') + ':00';
    hours.push(hour.getHours() + ':00');
    counts.push(call_count_by_hour[hourKey] || 0);
  }

  const data = {
    labels: hours,
    datasets: [
      {
        label: '调用次数',
        data: counts,
        backgroundColor: 'rgba(168, 85, 247, 0.5)',
        borderColor: 'rgba(168, 85, 247, 1)',
        borderWidth: 1,
      },
    ],
  };

  const options = {
    responsive: true,
    plugins: {
      legend: {
        display: false,
      },
    },
    scales: {
      y: {
        beginAtZero: true,
        title: {
          display: true,
          text: '调用次数',
        },
      },
    },
  };

  return (
    <ChartCard title="24小时调用统计" description="最近24小时的API调用次数分布">
      <div className="h-64">
        <Bar data={data} options={options} />
      </div>
    </ChartCard>
  );
};

interface CallRecordsTableProps {
  call_records: Array<{
    session_key: string;
    timestamp: string;
    success: boolean;
    latency: number;
  }>;
}

// 错误类型分布图表组件
interface ErrorTypeChartProps {
  error_types: Record<string, number>;
}

export const ErrorTypeChart: React.FC<ErrorTypeChartProps> = ({ error_types }) => {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const chartRef = useRef<Chart | null>(null);

  useEffect(() => {
    if (!canvasRef.current) return;

    const ctx = canvasRef.current.getContext('2d');
    if (!ctx) return;

    // 销毁现有图表
    if (chartRef.current) {
      chartRef.current.destroy();
    }

    // 错误类型映射
    const errorTypeLabels: Record<string, string> = {
      'rate_limit': '限流',
      'auth': '认证',
      'server': '服务器',
      'network': '网络',
      'timeout': '超时',
      'other': '其他'
    };

    const labels = Object.keys(error_types).map(key => errorTypeLabels[key] || key);
    const data = Object.values(error_types);
    const colors = [
      '#ef4444', // red - rate_limit
      '#f59e0b', // amber - auth  
      '#8b5cf6', // violet - server
      '#06b6d4', // cyan - network
      '#f97316', // orange - timeout
      '#6b7280'  // gray - other
    ];

    chartRef.current = new Chart(ctx, {
      type: 'doughnut',
      data: {
        labels: labels,
        datasets: [{
          data: data,
          backgroundColor: colors.slice(0, data.length),
          borderWidth: 2,
          borderColor: '#ffffff'
        }]
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          legend: {
            position: 'right',
            labels: {
              padding: 20,
              usePointStyle: true,
              font: { size: 12 }
            }
          },
          tooltip: {
            callbacks: {
              label: function(context) {
                const label = context.label || '';
                const value = context.parsed || 0;
                const total = context.dataset.data.reduce((a: number, b: number) => a + b, 0);
                const percentage = ((value / total) * 100).toFixed(1);
                return `${label}: ${value} (${percentage}%)`;
              }
            }
          }
        }
      }
    });

    return () => {
      if (chartRef.current) {
        chartRef.current.destroy();
      }
    };
  }, [error_types]);

  const totalErrors = Object.values(error_types).reduce((sum, count) => sum + count, 0);

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center space-x-2">
          <AlertTriangle className="w-5 h-5 text-red-600" />
          <span>错误类型分布</span>
        </CardTitle>
        <CardDescription>
          系统错误类型统计分布 (总计: {totalErrors} 个错误)
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="h-64">
          {totalErrors > 0 ? (
            <canvas ref={canvasRef} />
          ) : (
            <div className="flex items-center justify-center h-full text-muted-foreground">
              <div className="text-center">
                <CheckCircle className="w-12 h-12 mx-auto mb-2 text-green-600" />
                <p>暂无错误记录</p>
              </div>
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  );
};

// 熔断器状态可视化组件
interface CircuitBreakerChartProps {
  sessions: Array<{
    session_key: string;
    circuit_breaker?: {
      state: number;
      failure_count: number;
      threshold: number;
      timeout: number;
      next_attempt: string;
    };
  }>;
}

export const CircuitBreakerChart: React.FC<CircuitBreakerChartProps> = ({ sessions }) => {
  const circuitStates = sessions.map(session => {
    const cb = session.circuit_breaker;
    if (!cb) return { state: '未启用', stateNum: -1 };
    
    switch (cb.state) {
      case 0: return { state: '关闭', stateNum: 0, color: 'bg-green-500' };
      case 1: return { state: '开启', stateNum: 1, color: 'bg-red-500' };
      case 2: return { state: '半开', stateNum: 2, color: 'bg-yellow-500' };
      default: return { state: '未知', stateNum: -1, color: 'bg-gray-500' };
    }
  });

  const stateCounts = circuitStates.reduce((acc, state) => {
    acc[state.state] = (acc[state.state] || 0) + 1;
    return acc;
  }, {} as Record<string, number>);

  const enabledSessions = sessions.filter(s => s.circuit_breaker).length;

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center space-x-2">
          <Shield className="w-5 h-5 text-blue-600" />
          <span>熔断器状态</span>
        </CardTitle>
        <CardDescription>
          Session熔断器状态监控 (启用: {enabledSessions}/{sessions.length})
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          {/* 状态概览 */}
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <div className="text-center">
              <div className="w-3 h-3 bg-green-500 rounded-full mx-auto mb-1"></div>
              <div className="text-sm font-medium">关闭</div>
              <div className="text-xs text-muted-foreground">{stateCounts['关闭'] || 0}</div>
            </div>
            <div className="text-center">
              <div className="w-3 h-3 bg-red-500 rounded-full mx-auto mb-1"></div>
              <div className="text-sm font-medium">开启</div>
              <div className="text-xs text-muted-foreground">{stateCounts['开启'] || 0}</div>
            </div>
            <div className="text-center">
              <div className="w-3 h-3 bg-yellow-500 rounded-full mx-auto mb-1"></div>
              <div className="text-sm font-medium">半开</div>
              <div className="text-xs text-muted-foreground">{stateCounts['半开'] || 0}</div>
            </div>
            <div className="text-center">
              <div className="w-3 h-3 bg-gray-500 rounded-full mx-auto mb-1"></div>
              <div className="text-sm font-medium">未启用</div>
              <div className="text-xs text-muted-foreground">{stateCounts['未启用'] || 0}</div>
            </div>
          </div>

          {/* 详细状态列表 */}
          <div className="space-y-2 max-h-32 overflow-y-auto">
            {sessions.map((session, index) => {
              const cb = session.circuit_breaker;
              const state = circuitStates[index];
              
              return (
                <div key={session.session_key} className="flex items-center justify-between text-xs p-2 bg-muted/30 rounded">
                  <div className="flex items-center space-x-2">
                    <div className={`w-2 h-2 rounded-full ${state.color}`}></div>
                    <span className="font-mono">{session.session_key.substring(0, 8)}...</span>
                  </div>
                  <div className="flex items-center space-x-4 text-muted-foreground">
                    {cb && (
                      <>
                        <span>失败: {cb.failure_count}/{cb.threshold}</span>
                        <span>超时: {cb.timeout}s</span>
                      </>
                    )}
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      </CardContent>
    </Card>
  );
};

// 冷却时间倒计时组件
interface CooldownTimerProps {
  sessions: Array<{
    session_key: string;
    status: number;
    cooldown_until?: string;
  }>;
}

export const CooldownTimer: React.FC<CooldownTimerProps> = ({ sessions }) => {
  const [timeLeft, setTimeLeft] = React.useState<Record<string, string>>({});

  React.useEffect(() => {
    const timer = setInterval(() => {
      const now = new Date();
      const newTimeLeft: Record<string, string> = {};

      sessions.forEach(session => {
        if (session.cooldown_until) {
          const cooldownTime = new Date(session.cooldown_until);
          const diff = cooldownTime.getTime() - now.getTime();
          
          if (diff > 0) {
            const minutes = Math.floor(diff / 60000);
            const seconds = Math.floor((diff % 60000) / 1000);
            newTimeLeft[session.session_key] = `${minutes}:${seconds.toString().padStart(2, '0')}`;
          } else {
            newTimeLeft[session.session_key] = '已就绪';
          }
        }
      });

      setTimeLeft(newTimeLeft);
    }, 1000);

    return () => clearInterval(timer);
  }, [sessions]);

  const coolingSessions = sessions.filter(s => s.status === 1 && s.cooldown_until);

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center space-x-2">
          <Timer className="w-5 h-5 text-orange-600" />
          <span>冷却时间监控</span>
        </CardTitle>
        <CardDescription>
          Session冷却倒计时状态 (冷却中: {coolingSessions.length})
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-2">
          {coolingSessions.length === 0 ? (
            <div className="text-center py-4 text-muted-foreground">
              <CheckCircle className="w-8 h-8 mx-auto mb-2 text-green-600" />
              <p>暂无Session在冷却中</p>
            </div>
          ) : (
            coolingSessions.map(session => (
              <div key={session.session_key} className="flex items-center justify-between p-3 bg-muted/30 rounded-lg">
                <div className="flex items-center space-x-3">
                  <div className="w-2 h-2 bg-orange-500 rounded-full animate-pulse"></div>
                  <span className="font-mono text-sm">{session.session_key.substring(0, 8)}...</span>
                </div>
                <div className="flex items-center space-x-2">
                  <span className="text-sm font-medium text-orange-600">
                    {timeLeft[session.session_key] || '计算中...'}
                  </span>
                  <Timer className="w-4 h-4 text-orange-600" />
                </div>
              </div>
            ))
          )}
        </div>
      </CardContent>
    </Card>
  );
};

// 权重信息展示组件
interface WeightInfoProps {
  sessions: Array<{
    session_key: string;
    weights: number;
    health_score: number;
    success_count: number;
    total_requests: number;
  }>;
}

export const WeightInfo: React.FC<WeightInfoProps> = ({ sessions }) => {
  // 计算权重统计
  const totalWeight = sessions.reduce((sum, session) => sum + session.weights, 0);
  const avgWeight = sessions.length > 0 ? totalWeight / sessions.length : 0;
  
  // 按权重排序
  const sortedByWeight = [...sessions].sort((a, b) => b.weights - a.weights);

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center space-x-2">
          <Activity className="w-5 h-5 text-purple-600" />
          <span>权重信息</span>
        </CardTitle>
        <CardDescription>
          Session权重分配和性能统计 (总权重: {totalWeight.toFixed(2)}, 平均: {avgWeight.toFixed(2)})
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          {/* 权重分布概览 */}
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <div className="text-center">
              <div className="text-lg font-bold text-purple-600">{totalWeight.toFixed(2)}</div>
              <div className="text-xs text-muted-foreground">总权重</div>
            </div>
            <div className="text-center">
              <div className="text-lg font-bold text-blue-600">{avgWeight.toFixed(2)}</div>
              <div className="text-xs text-muted-foreground">平均权重</div>
            </div>
            <div className="text-center">
              <div className="text-lg font-bold text-green-600">
                {sortedByWeight[0]?.weights.toFixed(2) || '0.00'}
              </div>
              <div className="text-xs text-muted-foreground">最高权重</div>
            </div>
            <div className="text-center">
              <div className="text-lg font-bold text-orange-600">
                {sortedByWeight[sortedByWeight.length - 1]?.weights.toFixed(2) || '0.00'}
              </div>
              <div className="text-xs text-muted-foreground">最低权重</div>
            </div>
          </div>

          {/* 权重详细列表 */}
          <div className="space-y-3">
            <div className="text-sm font-medium text-muted-foreground mb-2">权重详细分布:</div>
            {sortedByWeight.map((session, index) => {
              const successRate = session.total_requests > 0 
                ? (session.success_count / session.total_requests) * 100 
                : 0;
              
              return (
                <div key={session.session_key} className="flex items-center justify-between p-3 bg-muted/30 rounded-lg">
                  <div className="flex items-center space-x-3">
                    <div className="flex items-center justify-center w-8 h-8 bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-200 rounded-full text-xs font-bold">
                      #{index + 1}
                    </div>
                    <div>
                      <div className="font-mono text-sm">{session.session_key.substring(0, 8)}...</div>
                      <div className="text-xs text-muted-foreground">
                        健康度: {(session.health_score * 100).toFixed(1)}% | 成功率: {successRate.toFixed(1)}%
                      </div>
                    </div>
                  </div>
                  <div className="flex items-center space-x-3">
                    <div className="text-right">
                      <div className="text-sm font-medium text-purple-600">
                        {session.weights.toFixed(3)}
                      </div>
                      <div className="text-xs text-muted-foreground">
                        {((session.weights / totalWeight) * 100).toFixed(1)}%
                      </div>
                    </div>
                    <div className="w-16 bg-gray-200 rounded-full h-2">
                      <div 
                        className="bg-purple-600 h-2 rounded-full transition-all duration-300"
                        style={{ 
                          width: `${(session.weights / Math.max(...sessions.map(s => s.weights))) * 100}%` 
                        }}
                      />
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      </CardContent>
    </Card>
  );
};

// 调度策略对比组件
interface StrategyComparisonProps {
  sessions: Array<{
    session_key: string;
    health_score: number;
    success_count: number;
    total_requests: number;
    avg_response_time: number;
    weights: number;
  }>;
  currentStrategy: string;
}

export const StrategyComparison: React.FC<StrategyComparisonProps> = ({ sessions, currentStrategy }) => {
  // 模拟不同策略的性能指标
  const calculateStrategyMetrics = () => {
    const strategies = [
      { name: 'round_robin', label: '轮询调度', color: '#3b82f6' },
      { name: 'health_priority', label: '健康优先', color: '#10b981' },
      { name: 'weighted', label: '权重调度', color: '#8b5cf6' },
      { name: 'adaptive', label: '自适应调度', color: '#f59e0b' }
    ];

    return strategies.map(strategy => {
      let metrics;
      
      switch (strategy.name) {
        case 'round_robin':
          metrics = {
            avgResponseTime: sessions.reduce((sum, s) => sum + s.avg_response_time, 0) / sessions.length,
            successRate: sessions.reduce((sum, s) => sum + (s.success_count / s.total_requests), 0) / sessions.length,
            loadDistribution: sessions.map(() => 100 / sessions.length) // 均匀分布
          };
          break;
          
        case 'health_priority':
          const healthWeighted = sessions.map(s => s.health_score);
          const totalHealth = healthWeighted.reduce((sum, w) => sum + w, 0);
          metrics = {
            avgResponseTime: sessions.reduce((sum, s, i) => sum + s.avg_response_time * healthWeighted[i], 0) / totalHealth,
            successRate: sessions.reduce((sum, s, i) => sum + (s.success_count / s.total_requests) * healthWeighted[i], 0) / totalHealth,
            loadDistribution: healthWeighted.map(w => (w / totalHealth) * 100)
          };
          break;
          
        case 'weighted':
          const sessionWeights = sessions.map(s => s.weights);
          const totalWeight = sessionWeights.reduce((sum, w) => sum + w, 0);
          metrics = {
            avgResponseTime: sessions.reduce((sum, s, i) => sum + s.avg_response_time * sessionWeights[i], 0) / totalWeight,
            successRate: sessions.reduce((sum, s, i) => sum + (s.success_count / s.total_requests) * sessionWeights[i], 0) / totalWeight,
            loadDistribution: sessionWeights.map(w => (w / totalWeight) * 100)
          };
          break;
          
        case 'adaptive':
          // 自适应策略：综合健康度和权重的智能调度
          const adaptiveScores = sessions.map(s => s.health_score * s.weights);
          const totalAdaptive = adaptiveScores.reduce((sum, score) => sum + score, 0);
          metrics = {
            avgResponseTime: sessions.reduce((sum, s, i) => sum + s.avg_response_time * adaptiveScores[i], 0) / totalAdaptive,
            successRate: sessions.reduce((sum, s, i) => sum + (s.success_count / s.total_requests) * adaptiveScores[i], 0) / totalAdaptive,
            loadDistribution: adaptiveScores.map(score => (score / totalAdaptive) * 100)
          };
          break;
          
        default:
          metrics = {
            avgResponseTime: 0,
            successRate: 0,
            loadDistribution: []
          };
      }

      return {
        ...strategy,
        ...metrics,
        score: (metrics.successRate * 100) + (1000 / Math.max(metrics.avgResponseTime, 1)) // 综合评分
      };
    });
  };

  const strategyMetrics = calculateStrategyMetrics();
  const bestStrategy = strategyMetrics.reduce((best, current) => 
    current.score > best.score ? current : best
  );

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center space-x-2">
          <Zap className="w-5 h-5 text-yellow-600" />
          <span>调度策略对比</span>
        </CardTitle>
        <CardDescription>
          不同调度策略的性能对比分析 (当前: {strategyMetrics.find(s => s.name === currentStrategy)?.label || '未知'})
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-6">
          {/* 策略性能概览 */}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            {strategyMetrics.map((strategy) => (
              <div 
                key={strategy.name}
                className={`p-4 rounded-lg border-2 transition-all duration-300 ${
                  strategy.name === currentStrategy 
                    ? 'border-blue-500 bg-blue-50 dark:bg-blue-950' 
                    : strategy.name === bestStrategy.name
                    ? 'border-green-500 bg-green-50 dark:bg-green-950'
                    : 'border-gray-200 dark:border-gray-700'
                }`}
              >
                <div className="flex items-center justify-between mb-2">
                  <div className="flex items-center space-x-2">
                    <div 
                      className="w-3 h-3 rounded-full" 
                      style={{ backgroundColor: strategy.color }}
                    ></div>
                    <span className="font-medium text-sm">{strategy.label}</span>
                  </div>
                  {strategy.name === currentStrategy && (
                    <span className="text-xs bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200 px-2 py-1 rounded">
                      当前
                    </span>
                  )}
                  {strategy.name === bestStrategy.name && strategy.name !== currentStrategy && (
                    <span className="text-xs bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200 px-2 py-1 rounded">
                      推荐
                    </span>
                  )}
                </div>
                
                <div className="space-y-1 text-xs">
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">响应时间:</span>
                    <span className="font-medium">{strategy.avgResponseTime.toFixed(0)}ms</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">成功率:</span>
                    <span className="font-medium">{(strategy.successRate * 100).toFixed(1)}%</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">综合评分:</span>
                    <span className="font-medium">{strategy.score.toFixed(1)}</span>
                  </div>
                </div>
              </div>
            ))}
          </div>

          {/* 负载分布对比 */}
          <div>
            <div className="text-sm font-medium text-muted-foreground mb-3">负载分布对比:</div>
            <div className="space-y-3">
              {sessions.slice(0, 5).map((session, index) => (
                <div key={session.session_key} className="space-y-2">
                  <div className="text-xs font-mono text-muted-foreground">
                    {session.session_key.substring(0, 8)}...
                  </div>
                  <div className="grid grid-cols-4 gap-2 h-6">
                    {strategyMetrics.map((strategy) => (
                      <div key={strategy.name} className="relative">
                        <div 
                          className="h-full rounded transition-all duration-300 flex items-center justify-center text-xs font-medium text-white"
                          style={{ 
                            width: `${Math.min(strategy.loadDistribution[index] || 0, 100)}%`,
                            backgroundColor: strategy.color 
                          }}
                        >
                          {(strategy.loadDistribution[index] || 0).toFixed(0)}%
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              ))}
            </div>
          </div>

          {/* 策略说明 */}
          <div className="text-xs text-muted-foreground bg-muted/50 p-3 rounded-lg">
            <div className="font-medium mb-1">策略说明:</div>
            <ul className="space-y-1">
              <li>• <span className="font-medium">轮询调度</span>: 按顺序均匀分配请求</li>
              <li>• <span className="font-medium">健康优先</span>: 优先选择健康度高的Session</li>
              <li>• <span className="font-medium">权重调度</span>: 根据权重比例分配请求</li>
              <li>• <span className="font-medium">自适应调度</span>: 智能综合健康度和权重进行调度</li>
            </ul>
          </div>
        </div>
      </CardContent>
    </Card>
  );
};

export const CallRecordsTable: React.FC<CallRecordsTableProps> = ({ call_records }) => {
  // 显示最近50条记录
  const recentRecords = call_records.slice(-50).reverse();
  
  const formatTime = (timestamp: string) => {
    return new Date(timestamp).toLocaleString('zh-CN');
  };

  const maskKey = (key: string) => {
    if (key.length <= 8) return key;
    return key.substring(0, 8) + '...' + key.substring(key.length - 4);
  };

  return (
    <ChartCard title="调用记录" description="最近50次API调用记录">
      <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b">
              <th className="text-left p-2">时间</th>
              <th className="text-left p-2">Session Key</th>
              <th className="text-left p-2">状态</th>
              <th className="text-left p-2">延迟</th>
            </tr>
          </thead>
          <tbody>
            {recentRecords.map((record, index) => (
              <tr key={index} className="border-b hover:bg-muted/50">
                <td className="p-2">{formatTime(record.timestamp)}</td>
                <td className="p-2 font-mono text-xs">{maskKey(record.session_key)}</td>
                <td className="p-2">
                  <span className={`px-2 py-1 rounded-full text-xs ${
                    record.success 
                      ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200'
                      : 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200'
                  }`}>
                    {record.success ? '成功' : '失败'}
                  </span>
                </td>
                <td className="p-2">
                  {record.latency > 0 ? `${record.latency}ms` : '-'}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </ChartCard>
  );
};

