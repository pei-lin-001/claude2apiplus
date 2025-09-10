import React, { useState, useEffect, useCallback } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import {
  BarChart3,
  Activity,
  Users,
  Zap,
  RefreshCw,
  AlertCircle,
  TrendingUp,
  Clock,
  CheckCircle,
  XCircle,
  Wifi,
  WifiOff,
  Globe,
  Heart,
  History,
  BarChart,
} from 'lucide-react';
import { SessionHealth, SystemStats } from '@/types';
import { SessionStatusChart, ResponseTimeChart, SuccessRateChart, HealthScoreChart, CallCountChart, CallRecordsTable, ErrorTypeChart, CircuitBreakerChart, CooldownTimer, WeightInfo, StrategyComparison } from '@/components/ui/charts';
import { useWebSocket } from '@/hooks/useWebSocket';

const StatisticsPage: React.FC = () => {
  const [stats, setStats] = useState<SystemStats | null>(null);
  const [sessions, setSessions] = useState<SessionHealth[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [lastUpdate, setLastUpdate] = useState<Date>(new Date());
  const [wsConnected, setWsConnected] = useState(false);
  const [currentStrategy, setCurrentStrategy] = useState<string>('round_robin');

  const handleWsMessage = useCallback((data: any) => {
    if (data.type === 'stats_update') setStats(data.stats);
    if (data.type === 'sessions_update') setSessions(data.sessions);
    setLastUpdate(new Date());
  }, []);

  useWebSocket({
    url: `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}/ws`,
    onMessage: handleWsMessage,
    onOpen: () => setWsConnected(true),
    onClose: () => setWsConnected(false),
    onError: () => setWsConnected(false),
  });

  const loadData = async () => {
    setLoading(true);
    setError(null);
    try {
      const apiKey = localStorage.getItem('apiKey') || 'test-api-key-123';
      const headers = { Authorization: `Bearer ${apiKey}` } as Record<string, string>;
      const [statsResponse, sessionsResponse, configResponse] = await Promise.all([
        fetch('/admin/stats', { headers }),
        fetch('/admin/sessions', { headers }),
        fetch('/admin/config', { headers }),
      ]);
      if (!statsResponse.ok || !sessionsResponse.ok || !configResponse.ok) throw new Error('Failed to fetch data');
      const [statsData, sessionsData, configData] = await Promise.all([
        statsResponse.json(), 
        sessionsResponse.json(), 
        configResponse.json()
      ]);
      setStats(statsData);
      setSessions(sessionsData);
      setCurrentStrategy(configData.sessionManager?.scheduleStrategy || 'round_robin');
      setLastUpdate(new Date());
    } catch (err) {
      console.error('Failed to load statistics:', err);
      setError('加载统计数据失败，请检查权限和连接: ' + (err as Error).message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, []);

  const getStatusColor = (status: number) => {
    switch (status) {
      case 0:
        return 'text-green-600 bg-green-50 dark:text-green-400 dark:bg-green-950';
      case 1:
        return 'text-yellow-600 bg-yellow-50 dark:text-yellow-400 dark:bg-yellow-950';
      case 2:
      case 3:
        return 'text-red-600 bg-red-50 dark:text-red-400 dark:bg-red-950';
      default:
        return 'text-gray-600 bg-gray-50 dark:text-gray-400 dark:bg-gray-950';
    }
  };

  const getStatusTextLocal = (status: number) => {
    switch (status) {
      case 0:
        return '活跃';
      case 1:
        return '冷却中';
      case 2:
        return '失败';
      case 3:
        return '熔断';
      default:
        return '未知';
    }
  };

  const formatHealthScore = (score: number) => (score * 100).toFixed(1);
  const formatResponseTime = (time: number) => `${time.toFixed(0)}ms`;

  const getSuccessRate = () => {
    if (!sessions.length) return 0;
    const totalRequests = sessions.reduce((sum, s) => sum + s.total_requests, 0);
    const totalSuccess = sessions.reduce((sum, s) => sum + s.success_count, 0);
    return totalRequests > 0 ? (totalSuccess / totalRequests) * 100 : 0;
  };

  const getAvgResponseTime = () => {
    if (!sessions.length) return 0;
    const totalTime = sessions.reduce((sum, s) => sum + s.avg_response_time * s.total_requests, 0);
    const totalRequests = sessions.reduce((sum, s) => sum + s.total_requests, 0);
    return totalRequests > 0 ? totalTime / totalRequests : 0;
  };

  if (loading) {
    return (
      <div className="max-w-7xl mx-auto p-6">
        <div className="flex items-center justify-between mb-8">
          <div className="flex items-center space-x-3">
            <div className="w-8 h-8 bg-blue-500 rounded-lg animate-pulse" />
            <div className="h-8 bg-gray-200 dark:bg-gray-700 rounded w-48 animate-pulse" />
          </div>
          <div className="flex items-center space-x-4">
            <div className="h-6 bg-gray-200 dark:bg-gray-700 rounded w-32 animate-pulse" />
            <div className="w-6 h-6 bg-gray-200 dark:bg-gray-700 rounded-full animate-spin" />
          </div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="max-w-7xl mx-auto p-6">
        <div className="text-center py-16">
          <AlertCircle className="w-16 h-16 text-red-500 mx-auto mb-4" />
          <h2 className="text-xl font-semibold mb-2">加载失败</h2>
          <p className="text-muted-foreground mb-6">{error}</p>
          <Button onClick={loadData} size="lg">
            <RefreshCw className="w-4 h-4 mr-2" />重试
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-7xl mx-auto p-6 space-y-8">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-3">
          <div className="w-8 h-8 bg-gradient-to-br from-blue-500 to-purple-600 rounded-lg flex items-center justify-center">
            <BarChart3 className="w-5 h-5 text-white" />
          </div>
          <div>
            <h1 className="text-2xl font-bold text-gray-900 dark:text-white">分析中心</h1>
            <p className="text-sm text-gray-500 dark:text-gray-400">系统性能与使用情况分析</p>
          </div>
        </div>
        <div className="flex items-center space-x-4">
          <div className="flex items-center space-x-2">
            <div className={`w-2 h-2 rounded-full ${wsConnected ? 'bg-green-500' : 'bg-red-500'}`} />
            <span className="text-sm text-gray-500 dark:text-gray-400">
              {wsConnected ? '实时连接' : '连接断开'}
            </span>
          </div>
          <Button onClick={loadData} variant="outline" size="sm">
            <RefreshCw className={`w-4 h-4 mr-2 ${loading ? 'animate-spin' : ''}`} />
            刷新
          </Button>
        </div>
      </div>

      {/* Key Metrics */}
      {stats && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          <Card className="border-0 shadow-sm hover:shadow-md transition-shadow">
            <CardContent className="p-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium text-gray-600 dark:text-gray-400">总会话数</p>
                  <p className="text-3xl font-bold text-gray-900 dark:text-white mt-1">
                    {stats.total_sessions}
                  </p>
                  <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                    活跃: {stats.active_sessions}
                  </p>
                </div>
                <div className="w-12 h-12 bg-blue-100 dark:bg-blue-900 rounded-lg flex items-center justify-center">
                  <Users className="w-6 h-6 text-blue-600 dark:text-blue-400" />
                </div>
              </div>
            </CardContent>
          </Card>

          <Card className="border-0 shadow-sm hover:shadow-md transition-shadow">
            <CardContent className="p-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium text-gray-600 dark:text-gray-400">总请求数</p>
                  <p className="text-3xl font-bold text-gray-900 dark:text-white mt-1">
                    {stats.total_requests.toLocaleString()}
                  </p>
                  <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                    成功率: {getSuccessRate().toFixed(1)}%
                  </p>
                </div>
                <div className="w-12 h-12 bg-green-100 dark:bg-green-900 rounded-lg flex items-center justify-center">
                  <Activity className="w-6 h-6 text-green-600 dark:text-green-400" />
                </div>
              </div>
            </CardContent>
          </Card>

          <Card className="border-0 shadow-sm hover:shadow-md transition-shadow">
            <CardContent className="p-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium text-gray-600 dark:text-gray-400">平均响应时间</p>
                  <p className="text-3xl font-bold text-gray-900 dark:text-white mt-1">
                    {formatResponseTime(getAvgResponseTime())}
                  </p>
                  <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                    基于所有会话
                  </p>
                </div>
                <div className="w-12 h-12 bg-purple-100 dark:bg-purple-900 rounded-lg flex items-center justify-center">
                  <Clock className="w-6 h-6 text-purple-600 dark:text-purple-400" />
                </div>
              </div>
            </CardContent>
          </Card>

          <Card className="border-0 shadow-sm hover:shadow-md transition-shadow">
            <CardContent className="p-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium text-gray-600 dark:text-gray-400">系统运行时间</p>
                  <p className="text-3xl font-bold text-gray-900 dark:text-white mt-1">
                    {stats.uptime}
                  </p>
                  <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                    稳定运行中
                  </p>
                </div>
                <div className="w-12 h-12 bg-orange-100 dark:bg-orange-900 rounded-lg flex items-center justify-center">
                  <Zap className="w-6 h-6 text-orange-600 dark:text-orange-400" />
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Main Content Tabs */}
      <Tabs defaultValue="performance" className="space-y-6">
        <TabsList className="grid w-full grid-cols-4 max-w-2xl">
          <TabsTrigger value="performance" className="flex items-center space-x-2">
            <BarChart className="w-4 h-4" />
            <span>性能分析</span>
          </TabsTrigger>
          <TabsTrigger value="sessions" className="flex items-center space-x-2">
            <Heart className="w-4 h-4" />
            <span>会话健康</span>
          </TabsTrigger>
          <TabsTrigger value="calls" className="flex items-center space-x-2">
            <History className="w-4 h-4" />
            <span>调用历史</span>
          </TabsTrigger>
          <TabsTrigger value="overview" className="flex items-center space-x-2">
            <Globe className="w-4 h-4" />
            <span>总览</span>
          </TabsTrigger>
        </TabsList>

        {/* Performance Analytics Tab */}
        <TabsContent value="performance" className="space-y-6">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <Card className="border-0 shadow-sm">
              <CardHeader>
                <CardTitle className="flex items-center space-x-2">
                  <TrendingUp className="w-5 h-5" />
                  <span>响应时间趋势</span>
                </CardTitle>
                <CardDescription>会话平均响应时间分析</CardDescription>
              </CardHeader>
              <CardContent>
                <ResponseTimeChart sessions={sessions} />
              </CardContent>
            </Card>

            <Card className="border-0 shadow-sm">
              <CardHeader>
                <CardTitle className="flex items-center space-x-2">
                  <CheckCircle className="w-5 h-5" />
                  <span>成功率分析</span>
                </CardTitle>
                <CardDescription>各会话请求成功率对比</CardDescription>
              </CardHeader>
              <CardContent>
                <SuccessRateChart sessions={sessions} />
              </CardContent>
            </Card>

            <Card className="border-0 shadow-sm">
              <CardHeader>
                <CardTitle className="flex items-center space-x-2">
                  <BarChart3 className="w-5 h-5" />
                  <span>调用次数分布</span>
                </CardTitle>
                <CardDescription>按小时统计的调用次数</CardDescription>
              </CardHeader>
              <CardContent>
                <CallCountChart call_count_by_hour={stats?.call_count_by_hour || {}} />
              </CardContent>
            </Card>

            <Card className="border-0 shadow-sm">
              <CardHeader>
                <CardTitle className="flex items-center space-x-2">
                  <XCircle className="w-5 h-5" />
                  <span>错误类型分析</span>
                </CardTitle>
                <CardDescription>错误类型分布统计</CardDescription>
              </CardHeader>
              <CardContent>
                <ErrorTypeChart error_types={stats?.errors_by_type || {}} />
              </CardContent>
            </Card>
          </div>

          <Card className="border-0 shadow-sm">
            <CardHeader>
              <CardTitle className="flex items-center space-x-2">
                <Activity className="w-5 h-5" />
                <span>调度策略对比</span>
              </CardTitle>
              <CardDescription>不同调度策略的性能对比</CardDescription>
            </CardHeader>
            <CardContent>
              <StrategyComparison sessions={sessions} currentStrategy={currentStrategy} />
            </CardContent>
          </Card>
        </TabsContent>

        {/* Session Health Tab */}
        <TabsContent value="sessions" className="space-y-6">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <Card className="border-0 shadow-sm">
              <CardHeader>
                <CardTitle className="flex items-center space-x-2">
                  <Heart className="w-5 h-5" />
                  <span>会话健康状态</span>
                </CardTitle>
                <CardDescription>各会话的健康状态分布</CardDescription>
              </CardHeader>
              <CardContent>
                <SessionStatusChart sessions={sessions} />
              </CardContent>
            </Card>

            <Card className="border-0 shadow-sm">
              <CardHeader>
                <CardTitle className="flex items-center space-x-2">
                  <Activity className="w-5 h-5" />
                  <span>健康分数趋势</span>
                </CardTitle>
                <CardDescription>会话健康分数对比</CardDescription>
              </CardHeader>
              <CardContent>
                <HealthScoreChart sessions={sessions} />
              </CardContent>
            </Card>

            <Card className="border-0 shadow-sm">
              <CardHeader>
                <CardTitle className="flex items-center space-x-2">
                  <XCircle className="w-5 h-5" />
                  <span>熔断器状态</span>
                </CardTitle>
                <CardDescription>会话熔断器状态监控</CardDescription>
              </CardHeader>
              <CardContent>
                <CircuitBreakerChart sessions={sessions} />
              </CardContent>
            </Card>

            <Card className="border-0 shadow-sm">
              <CardHeader>
                <CardTitle className="flex items-center space-x-2">
                  <Clock className="w-5 h-5" />
                  <span>冷却计时器</span>
                </CardTitle>
                <CardDescription>会话冷却时间状态</CardDescription>
              </CardHeader>
              <CardContent>
                <CooldownTimer sessions={sessions} />
              </CardContent>
            </Card>

            <Card className="border-0 shadow-sm">
              <CardHeader>
                <CardTitle className="flex items-center space-x-2">
                  <BarChart3 className="w-5 h-5" />
                  <span>权重信息</span>
                </CardTitle>
                <CardDescription>会话权重分布</CardDescription>
              </CardHeader>
              <CardContent>
                <WeightInfo sessions={sessions} />
              </CardContent>
            </Card>
          </div>

          <Card className="border-0 shadow-sm">
            <CardHeader>
              <CardTitle>会话状态详情</CardTitle>
              <CardDescription>各个会话的实时健康状态和性能指标</CardDescription>
            </CardHeader>
            <CardContent>
              {sessions.length === 0 ? (
                <div className="text-center py-8">
                  <Users className="w-12 h-12 mx-auto text-muted-foreground mb-4" />
                  <p className="text-muted-foreground">暂无会话数据</p>
                </div>
              ) : (
                <div className="space-y-4">
                  {sessions.map((session) => (
                    <div key={session.session_key} className="border border-gray-200 dark:border-gray-700 rounded-lg p-4 hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors">
                      <div className="flex items-start justify-between">
                        <div className="flex-1">
                          <div className="flex items-center space-x-3 mb-2">
                            <h3 className="font-mono text-sm bg-gray-100 dark:bg-gray-800 px-2 py-1 rounded">
                              {session.session_key.substring(0, 12)}...
                            </h3>
                            <span className={`px-2 py-1 rounded-full text-xs font-medium ${getStatusColor(session.status)}`}>
                              {getStatusTextLocal(session.status)}
                            </span>
                          </div>

                          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
                            <div>
                              <span className="text-gray-500 dark:text-gray-400">健康分数:</span>
                              <span className="ml-1 font-medium">{formatHealthScore(session.health_score)}</span>
                            </div>
                            <div>
                              <span className="text-gray-500 dark:text-gray-400">成功率:</span>
                              <span className="ml-1 font-medium">
                                {session.total_requests > 0 ? ((session.success_count / session.total_requests) * 100).toFixed(1) : 0}%
                              </span>
                            </div>
                            <div>
                              <span className="text-gray-500 dark:text-gray-400">平均响应:</span>
                              <span className="ml-1 font-medium">{formatResponseTime(session.avg_response_time)}</span>
                            </div>
                            <div>
                              <span className="text-gray-500 dark:text-gray-400">总请求:</span>
                              <span className="ml-1 font-medium">{session.total_requests}</span>
                            </div>
                          </div>

                          <div className="mt-3">
                            <div className="flex items-center justify-between text-xs text-gray-500 dark:text-gray-400 mb-1">
                              <span>健康度</span>
                              <span>{formatHealthScore(session.health_score)}%</span>
                            </div>
                            <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
                              <div
                                className={`h-2 rounded-full transition-all duration-300 ${
                                  session.health_score >= 0.8 ? 'bg-green-500' : 
                                  session.health_score >= 0.5 ? 'bg-yellow-500' : 'bg-red-500'
                                }`}
                                style={{ width: `${session.health_score * 100}%` }}
                              />
                            </div>
                          </div>
                        </div>

                        <div className="text-right text-xs text-gray-500 dark:text-gray-400">
                          <div>最后使用</div>
                          <div>{new Date(session.last_used).toLocaleString('zh-CN')}</div>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        {/* Call History Tab */}
        <TabsContent value="calls" className="space-y-6">
          {stats && (
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              <Card className="border-0 shadow-sm">
                <CardHeader>
                  <CardTitle className="flex items-center space-x-2">
                    <History className="w-5 h-5" />
                    <span>调用记录</span>
                  </CardTitle>
                  <CardDescription>最近的API调用记录</CardDescription>
                </CardHeader>
                <CardContent>
                  <CallRecordsTable call_records={stats.call_records || []} />
                </CardContent>
              </Card>

              <Card className="border-0 shadow-sm">
                <CardHeader>
                  <CardTitle className="flex items-center space-x-2">
                    <BarChart3 className="w-5 h-5" />
                    <span>调用次数统计</span>
                  </CardTitle>
                  <CardDescription>按小时统计的调用次数</CardDescription>
                </CardHeader>
                <CardContent>
                  <CallCountChart call_count_by_hour={stats.call_count_by_hour || {}} />
                </CardContent>
              </Card>
            </div>
          )}
        </TabsContent>

        {/* Overview Tab */}
        <TabsContent value="overview" className="space-y-6">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <Card className="border-0 shadow-sm">
              <CardHeader>
                <CardTitle className="flex items-center space-x-2">
                  <Users className="w-5 h-5" />
                  <span>会话状态分布</span>
                </CardTitle>
                <CardDescription>各状态会话数量统计</CardDescription>
              </CardHeader>
              <CardContent>
                <div className="grid grid-cols-2 gap-4">
                  <div className="text-center p-4 bg-green-50 dark:bg-green-900 rounded-lg">
                    <div className="text-2xl font-bold text-green-600 dark:text-green-400">
                      {sessions.filter((s) => s.status === 0).length}
                    </div>
                    <div className="text-sm text-green-600 dark:text-green-400">活跃</div>
                  </div>
                  <div className="text-center p-4 bg-yellow-50 dark:bg-yellow-900 rounded-lg">
                    <div className="text-2xl font-bold text-yellow-600 dark:text-yellow-400">
                      {sessions.filter((s) => s.status === 1).length}
                    </div>
                    <div className="text-sm text-yellow-600 dark:text-yellow-400">冷却中</div>
                  </div>
                  <div className="text-center p-4 bg-red-50 dark:bg-red-900 rounded-lg">
                    <div className="text-2xl font-bold text-red-600 dark:text-red-400">
                      {sessions.filter((s) => s.status === 2).length}
                    </div>
                    <div className="text-sm text-red-600 dark:text-red-400">失败</div>
                  </div>
                  <div className="text-center p-4 bg-red-50 dark:bg-red-900 rounded-lg">
                    <div className="text-2xl font-bold text-red-600 dark:text-red-400">
                      {sessions.filter((s) => s.status === 3).length}
                    </div>
                    <div className="text-sm text-red-600 dark:text-red-400">熔断</div>
                  </div>
                </div>
              </CardContent>
            </Card>

            <Card className="border-0 shadow-sm">
              <CardHeader>
                <CardTitle className="flex items-center space-x-2">
                  <Activity className="w-5 h-5" />
                  <span>系统信息</span>
                </CardTitle>
                <CardDescription>系统运行状态信息</CardDescription>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  <div className="flex justify-between">
                    <span className="text-sm text-gray-500 dark:text-gray-400">当前策略</span>
                    <span className="text-sm font-medium">{currentStrategy}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-sm text-gray-500 dark:text-gray-400">WebSocket</span>
                    <span className={`text-sm font-medium ${wsConnected ? 'text-green-600' : 'text-red-600'}`}>
                      {wsConnected ? '已连接' : '未连接'}
                    </span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-sm text-gray-500 dark:text-gray-400">最后更新</span>
                    <span className="text-sm font-medium">{lastUpdate.toLocaleTimeString('zh-CN')}</span>
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>
        </TabsContent>
      </Tabs>
    </div>
  );
};

export default StatisticsPage;