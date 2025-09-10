import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { 
  Key, 
  Plus, 
  Trash2, 
  Activity, 
  RefreshCw, 
  AlertCircle, 
  Edit, 
  Save, 
  X,
  CheckCircle,
  Zap,
  Users
} from 'lucide-react';
import { SessionCardSkeleton } from '@/components/ui/skeleton';

interface SessionInfo {
  sessionKey: string;
  orgID: string;
  isActive: boolean;
  healthScore?: number;
  status?: string;
  lastUsed?: string;
  errorCount?: number;
  successCount?: number;
  totalRequests?: number;
  avgResponseTime?: number;
}

const SessionManager: React.FC = () => {
  const [sessions, setSessions] = useState<SessionInfo[]>([]);
  const [newSessionKey, setNewSessionKey] = useState('');
  const [newOrgID, setNewOrgID] = useState('');
  const [loading, setLoading] = useState(false);
  const [initialLoading, setInitialLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [editingSession, setEditingSession] = useState<string | null>(null);
  const [editOrgID, setEditOrgID] = useState('');

  const loadSessions = async () => {
    if (initialLoading) {
      setLoading(true);
    }
    setError(null);
    
    try {
      const response = await fetch('/admin/sessions', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('apiKey') || 'test-api-key-123'}`,
        },
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const data = await response.json();
      
      const formattedSessions = data.map((session: any) => ({
        sessionKey: session.session_key || session.sessionKey,
        orgID: session.org_id || session.orgID || '',
        isActive: session.status === 'active',
        healthScore: session.health_score || 1.0,
        status: session.status || 'unknown',
        lastUsed: session.last_used ? new Date(session.last_used).toLocaleString() : '从未使用',
        errorCount: session.error_count || 0,
        successCount: session.success_count || 0,
        totalRequests: session.total_requests || 0,
        avgResponseTime: session.avg_response_time || 0
      }));
      
      setSessions(formattedSessions);
    } catch (err) {
      console.error('Failed to load sessions:', err);
      setError('加载会话失败: ' + (err as Error).message);
    } finally {
      setLoading(false);
      setInitialLoading(false);
    }
  };

  useEffect(() => {
    loadSessions();
  }, []);

  const handleAddSession = async () => {
    if (!newSessionKey.trim()) {
      setError('请输入Session Key');
      return;
    }

    setLoading(true);
    setError(null);
    setSuccess(null);

    try {
      const response = await fetch('/admin/sessions', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('apiKey') || 'test-api-key-123'}`,
        },
        body: JSON.stringify({
          sessionKey: newSessionKey.trim(),
          orgID: newOrgID.trim() || undefined,
        }),
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      await response.json();
      await loadSessions();
      
      setNewSessionKey('');
      setNewOrgID('');
      setSuccess('Session添加成功');
      
      setTimeout(() => setSuccess(null), 3000);
    } catch (err) {
      console.error('Failed to add session:', err);
      setError('添加Session失败: ' + (err as Error).message);
    } finally {
      setLoading(false);
    }
  };

  const handleDeleteSession = async (sessionKey: string) => {
    if (!window.confirm('确定要删除这个Session吗？')) {
      return;
    }

    setLoading(true);
    setError(null);

    try {
      const response = await fetch(`/admin/sessions/${encodeURIComponent(sessionKey)}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('apiKey') || 'test-api-key-123'}`,
        },
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      await loadSessions();
      setSuccess('Session删除成功');
      
      setTimeout(() => setSuccess(null), 3000);
    } catch (err) {
      console.error('Failed to delete session:', err);
      setError('删除Session失败: ' + (err as Error).message);
    } finally {
      setLoading(false);
    }
  };

  const handleEditSession = (sessionKey: string, currentOrgID: string) => {
    setEditingSession(sessionKey);
    setEditOrgID(currentOrgID);
  };

  const handleCancelEdit = () => {
    setEditingSession(null);
    setEditOrgID('');
  };

  const handleSaveSession = async (sessionKey: string) => {
    setLoading(true);
    setError(null);

    try {
      const response = await fetch(`/admin/sessions/${encodeURIComponent(sessionKey)}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('apiKey') || 'test-api-key-123'}`,
        },
        body: JSON.stringify({
          orgID: editOrgID.trim(),
        }),
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      await loadSessions();
      setEditingSession(null);
      setEditOrgID('');
      setSuccess('Session更新成功');
      
      setTimeout(() => setSuccess(null), 3000);
    } catch (err) {
      console.error('Failed to update session:', err);
      setError('更新Session失败: ' + (err as Error).message);
    } finally {
      setLoading(false);
    }
  };

  const maskSessionKey = (key: string) => {
    if (key.length <= 8) return key;
    return key.substring(0, 8) + '...' + key.substring(key.length - 4);
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'active':
        return <Badge className="bg-green-100 text-green-800 border-green-200">活跃</Badge>;
      case 'cooling':
        return <Badge className="bg-yellow-100 text-yellow-800 border-yellow-200">冷却中</Badge>;
      case 'unknown':
        return <Badge className="bg-gray-100 text-gray-800 border-gray-200">未知</Badge>;
      case 'failed':
        return <Badge className="bg-red-100 text-red-800 border-red-200">失败</Badge>;
      case 'circuit_open':
        return <Badge className="bg-red-100 text-red-800 border-red-200">熔断</Badge>;
      default:
        return <Badge className="bg-gray-100 text-gray-800 border-gray-200">未知</Badge>;
    }
  };

  const getHealthColor = (score: number) => {
    if (score >= 0.8) return 'text-green-600';
    if (score >= 0.6) return 'text-yellow-600';
    return 'text-red-600';
  };

  return (
    <div className="max-w-6xl mx-auto p-6 space-y-8">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-3">
          <div className="w-8 h-8 bg-gradient-to-br from-emerald-500 to-teal-600 rounded-lg flex items-center justify-center">
            <Key className="w-5 h-5 text-white" />
          </div>
          <div>
            <h1 className="text-2xl font-bold text-gray-900 dark:text-white">会话管理</h1>
            <p className="text-sm text-gray-500 dark:text-gray-400">管理Claude API会话密钥和配置</p>
          </div>
        </div>
        <Button onClick={loadSessions} variant="outline" size="sm">
          <RefreshCw className={`w-4 h-4 mr-2 ${loading ? 'animate-spin' : ''}`} />
          刷新
        </Button>
      </div>

      {/* Stats Overview */}
      {sessions.length > 0 && (
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <Card className="border-0 shadow-sm">
            <CardContent className="p-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium text-gray-600 dark:text-gray-400">总会话数</p>
                  <p className="text-2xl font-bold text-gray-900 dark:text-white">{sessions.length}</p>
                </div>
                <Users className="w-8 h-8 text-blue-500" />
              </div>
            </CardContent>
          </Card>

          <Card className="border-0 shadow-sm">
            <CardContent className="p-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium text-gray-600 dark:text-gray-400">活跃会话</p>
                  <p className="text-2xl font-bold text-green-600 dark:text-green-400">
                    {sessions.filter(s => s.status === 'active').length}
                  </p>
                </div>
                <CheckCircle className="w-8 h-8 text-green-500" />
              </div>
            </CardContent>
          </Card>

          <Card className="border-0 shadow-sm">
            <CardContent className="p-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium text-gray-600 dark:text-gray-400">总请求数</p>
                  <p className="text-2xl font-bold text-gray-900 dark:text-white">
                    {sessions.reduce((sum, s) => sum + (s.totalRequests || 0), 0).toLocaleString()}
                  </p>
                </div>
                <Activity className="w-8 h-8 text-purple-500" />
              </div>
            </CardContent>
          </Card>

          <Card className="border-0 shadow-sm">
            <CardContent className="p-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium text-gray-600 dark:text-gray-400">平均健康度</p>
                  <p className="text-2xl font-bold text-gray-900 dark:text-white">
                    {((sessions.reduce((sum, s) => sum + (s.healthScore || 0), 0) / sessions.length) * 100).toFixed(0)}%
                  </p>
                </div>
                <Zap className="w-8 h-8 text-orange-500" />
              </div>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Add Session Form */}
      <Card className="border-0 shadow-sm">
        <CardHeader>
          <CardTitle className="flex items-center space-x-2">
            <Plus className="w-5 h-5" />
            <span>添加新会话</span>
          </CardTitle>
          <CardDescription>
            添加新的Claude API Session Key到会话池中
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="space-y-2">
              <label className="text-sm font-medium text-gray-700 dark:text-gray-300">Session Key</label>
              <Input
                placeholder="sk-ant-sid01-xxxxxxxxxx"
                value={newSessionKey}
                onChange={(e) => setNewSessionKey(e.target.value)}
                disabled={loading}
                className="transition-all duration-200"
              />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium text-gray-700 dark:text-gray-300">Organization ID (可选)</label>
              <Input
                placeholder="org-id"
                value={newOrgID}
                onChange={(e) => setNewOrgID(e.target.value)}
                disabled={loading}
                className="transition-all duration-200"
              />
            </div>
          </div>
          
          <Button 
            onClick={handleAddSession} 
            disabled={loading || !newSessionKey.trim()}
            className="w-full md:w-auto"
          >
            <Plus className="w-4 h-4 mr-2" />
            添加会话
          </Button>

          {error && (
            <div className="p-4 rounded-lg bg-red-50 border border-red-200 text-red-700 dark:bg-red-900/20 dark:border-red-800 dark:text-red-300 text-sm flex items-center space-x-2">
              <AlertCircle className="w-4 h-4" />
              <span>{error}</span>
            </div>
          )}

          {success && (
            <div className="p-4 rounded-lg bg-green-50 border border-green-200 text-green-700 dark:bg-green-900/20 dark:border-green-800 dark:text-green-300 text-sm">
              {success}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Sessions List */}
      <Card className="border-0 shadow-sm">
        <CardHeader>
          <CardTitle className="flex items-center space-x-2">
            <Activity className="w-5 h-5" />
            <span>会话列表</span>
          </CardTitle>
          <CardDescription>
            当前可用的Claude API会话列表
          </CardDescription>
        </CardHeader>
        <CardContent>
          {initialLoading ? (
            <div className="space-y-4">
              <div className="text-sm text-gray-500 dark:text-gray-400 mb-4">
                加载会话列表...
              </div>
              {[...Array(3)].map((_, i) => (
                <SessionCardSkeleton key={i} />
              ))}
            </div>
          ) : sessions.length === 0 ? (
            <div className="text-center py-12">
              <Key className="w-16 h-16 mx-auto text-gray-300 dark:text-gray-600 mb-4" />
              <p className="text-gray-500 dark:text-gray-400 text-lg">暂无会话配置</p>
              <p className="text-gray-400 dark:text-gray-500 mt-2">
                请添加Session Key以开始使用
              </p>
            </div>
          ) : (
            <div className="space-y-4">
              <div className="text-sm text-gray-500 dark:text-gray-400 mb-4">
                总计 {sessions.length} 个会话
              </div>
              
              {sessions.map((session, index) => (
                <div
                  key={session.sessionKey}
                  className="border border-gray-200 dark:border-gray-700 rounded-lg p-6 hover:shadow-md transition-all duration-200 bg-white dark:bg-gray-800"
                >
                  <div className="flex items-start justify-between">
                    <div className="flex-1 space-y-4">
                      {/* Header */}
                      <div className="flex items-center justify-between">
                        <div className="flex items-center space-x-3">
                          <span className="text-sm font-medium text-gray-500 dark:text-gray-400">
                            #{index + 1}
                          </span>
                          <h3 className="font-mono text-sm bg-gray-100 dark:bg-gray-700 px-3 py-1 rounded-md">
                            {maskSessionKey(session.sessionKey)}
                          </h3>
                          {getStatusBadge(session.status || 'unknown')}
                        </div>
                      </div>

                      {/* Stats Grid */}
                      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                        <div className="text-center p-3 bg-gray-50 dark:bg-gray-700 rounded-lg">
                          <div className={`text-lg font-bold ${getHealthColor(session.healthScore || 0)}`}>
                            {Math.round((session.healthScore || 0) * 100)}%
                          </div>
                          <div className="text-xs text-gray-500 dark:text-gray-400">健康度</div>
                        </div>
                        <div className="text-center p-3 bg-gray-50 dark:bg-gray-700 rounded-lg">
                          <div className="text-lg font-bold text-gray-900 dark:text-white">
                            {session.totalRequests || 0}
                          </div>
                          <div className="text-xs text-gray-500 dark:text-gray-400">总请求</div>
                        </div>
                        <div className="text-center p-3 bg-gray-50 dark:bg-gray-700 rounded-lg">
                          <div className="text-lg font-bold text-green-600 dark:text-green-400">
                            {session.successCount || 0}
                          </div>
                          <div className="text-xs text-gray-500 dark:text-gray-400">成功</div>
                        </div>
                        <div className="text-center p-3 bg-gray-50 dark:bg-gray-700 rounded-lg">
                          <div className="text-lg font-bold text-red-600 dark:text-red-400">
                            {session.errorCount || 0}
                          </div>
                          <div className="text-xs text-gray-500 dark:text-gray-400">错误</div>
                        </div>
                      </div>

                      {/* Org ID Section */}
                      {editingSession === session.sessionKey ? (
                        <div className="flex items-center space-x-2">
                          <Input
                            placeholder="Organization ID"
                            value={editOrgID}
                            onChange={(e) => setEditOrgID(e.target.value)}
                            className="flex-1"
                            disabled={loading}
                          />
                          <Button
                            size="sm"
                            onClick={() => handleSaveSession(session.sessionKey)}
                            disabled={loading}
                          >
                            <Save className="w-4 h-4" />
                          </Button>
                          <Button
                            size="sm"
                            variant="outline"
                            onClick={handleCancelEdit}
                            disabled={loading}
                          >
                            <X className="w-4 h-4" />
                          </Button>
                        </div>
                      ) : (
                        <div className="flex items-center justify-between">
                          <div>
                            {session.orgID ? (
                              <div className="text-sm text-gray-600 dark:text-gray-400">
                                Organization ID: <span className="font-medium">{session.orgID}</span>
                              </div>
                            ) : (
                              <div className="text-sm text-gray-400 dark:text-gray-500">
                                未设置 Organization ID
                              </div>
                            )}
                          </div>
                          <div className="flex items-center space-x-2">
                            <Button
                              variant="outline"
                              size="sm"
                              onClick={() => handleEditSession(session.sessionKey, session.orgID)}
                              disabled={loading}
                            >
                              <Edit className="w-4 h-4" />
                            </Button>
                            <Button
                              variant="destructive"
                              size="sm"
                              onClick={() => handleDeleteSession(session.sessionKey)}
                              disabled={loading}
                            >
                              <Trash2 className="w-4 h-4" />
                            </Button>
                          </div>
                        </div>
                      )}

                      {/* Footer */}
                      <div className="flex items-center justify-between text-xs text-gray-500 dark:text-gray-400 pt-2 border-t border-gray-200 dark:border-gray-700">
                        <div className="flex items-center space-x-4">
                          <span>平均响应: {session.avgResponseTime?.toFixed(0) || 0}ms</span>
                          <span>成功率: {session.totalRequests ? ((session.successCount || 0) / session.totalRequests * 100).toFixed(1) : 0}%</span>
                        </div>
                        <div className="text-right">
                          <div>最后使用</div>
                          <div>{session.lastUsed}</div>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
};

export default SessionManager;