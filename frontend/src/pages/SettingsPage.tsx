import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Switch } from '@/components/ui/switch';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import { useChatStore } from '@/stores/chat';
import { 
  Save, 
  Settings, 
  RefreshCw, 
  AlertCircle, 
  Check, 
  RotateCcw,
  Key,
  Moon,
  Sun,
  Zap,
  Shield,
  Database,
  Globe,
  Sliders,
  Activity
} from 'lucide-react';

interface SystemConfig {
  sessionManagerEnabled: boolean;
  totalSessions: number;
  retryCount: number;
  chatDelete: boolean;
  maxChatHistoryLength: number;
  noRolePrefix: boolean;
  promptDisableArtifacts: boolean;
  enableMirrorApi: boolean;
  mirrorApiPrefix: string;
  sessionManager?: {
    scheduleStrategy: string;
    maxRetryAttempts: number;
    healthCheckInterval: number;
    minHealthScore: number;
    circuitBreakerEnabled: boolean;
  };
}

const SettingsPage: React.FC = () => {
  const { darkMode, setDarkMode } = useChatStore();
  const [apiKey, setApiKey] = useState(localStorage.getItem('apiKey') || '');
  const [saveStatus, setSaveStatus] = useState<'idle' | 'saving' | 'success' | 'error'>('idle');
  const [config, setConfig] = useState<SystemConfig | null>(null);
  const [configLoading, setConfigLoading] = useState(true);
  const [configError, setConfigError] = useState<string | null>(null);
  const [editingConfig, setEditingConfig] = useState<SystemConfig | null>(null);
  const [saveConfigStatus, setSaveConfigStatus] = useState<'idle' | 'saving' | 'success' | 'error'>('idle');

  const loadConfig = async () => {
    setConfigLoading(true);
    setConfigError(null);
    
    try {
      const apiKey = localStorage.getItem('apiKey') || 'test-api-key-123';
      const response = await fetch('/admin/config', {
        headers: {
          'Authorization': `Bearer ${apiKey}`,
        },
      });

      if (!response.ok) {
        throw new Error('Failed to load configuration');
      }

      const configData = await response.json();
      setConfig(configData);
      setEditingConfig(configData);
    } catch (error) {
      console.error('Failed to load config:', error);
      setConfigError('加载配置失败: ' + (error as Error).message);
    } finally {
      setConfigLoading(false);
    }
  };

  const handleSaveConfig = async () => {
    if (!editingConfig) return;
    
    setSaveConfigStatus('saving');
    try {
      const apiKey = localStorage.getItem('apiKey') || 'test-api-key-123';
      const response = await fetch('/admin/config', {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${apiKey}`,
        },
        body: JSON.stringify(editingConfig),
      });

      if (!response.ok) {
        throw new Error('Failed to save configuration');
      }

      setConfig(editingConfig);
      setSaveConfigStatus('success');
      setTimeout(() => setSaveConfigStatus('idle'), 2000);
    } catch (error) {
      console.error('Failed to save config:', error);
      setSaveConfigStatus('error');
      setTimeout(() => setSaveConfigStatus('idle'), 2000);
    }
  };

  const handleResetConfig = () => {
    if (config) {
      setEditingConfig(config);
    }
  };

  useEffect(() => {
    loadConfig();
  }, []);

  const handleSaveApiKey = () => {
    setSaveStatus('saving');
    try {
      localStorage.setItem('apiKey', apiKey);
      setSaveStatus('success');
      setTimeout(() => setSaveStatus('idle'), 2000);
    } catch (error) {
      setSaveStatus('error');
      setTimeout(() => setSaveStatus('idle'), 2000);
    }
  };

  const handleToggleDarkMode = () => {
    setDarkMode(!darkMode);
  };

  return (
    <div className="max-w-6xl mx-auto p-6 space-y-8">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-3">
          <div className="w-8 h-8 bg-gradient-to-br from-indigo-500 to-purple-600 rounded-lg flex items-center justify-center">
            <Settings className="w-5 h-5 text-white" />
          </div>
          <div>
            <h1 className="text-2xl font-bold text-gray-900 dark:text-white">系统设置</h1>
            <p className="text-sm text-gray-500 dark:text-gray-400">配置系统参数和用户偏好</p>
          </div>
        </div>
        <Button onClick={loadConfig} variant="outline" size="sm">
          <RefreshCw className={`w-4 h-4 mr-2 ${configLoading ? 'animate-spin' : ''}`} />
          刷新配置
        </Button>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
        {/* API Configuration */}
        <Card className="border-0 shadow-sm">
          <CardHeader>
            <CardTitle className="flex items-center space-x-2">
              <Key className="w-5 h-5" />
              <span>API配置</span>
            </CardTitle>
            <CardDescription>
              配置Claude API的访问密钥
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="apiKey" className="text-sm font-medium text-gray-700 dark:text-gray-300">
                API密钥
              </Label>
              <Input
                id="apiKey"
                type="password"
                placeholder="sk-ant-sid01-xxxxxxxxxx"
                value={apiKey}
                onChange={(e) => setApiKey(e.target.value)}
                className="transition-all duration-200"
              />
            </div>
            
            <Button 
              onClick={handleSaveApiKey} 
              disabled={saveStatus === 'saving'}
              className="w-full"
            >
              <Save className="w-4 h-4 mr-2" />
              {saveStatus === 'saving' ? '保存中...' : '保存密钥'}
            </Button>

            {saveStatus === 'success' && (
              <div className="flex items-center space-x-2 text-sm text-green-600">
                <Check className="w-4 h-4" />
                <span>API密钥保存成功</span>
              </div>
            )}
            
            {saveStatus === 'error' && (
              <div className="flex items-center space-x-2 text-sm text-red-600">
                <AlertCircle className="w-4 h-4" />
                <span>保存失败，请重试</span>
              </div>
            )}
          </CardContent>
        </Card>

        {/* Interface Settings */}
        <Card className="border-0 shadow-sm">
          <CardHeader>
            <CardTitle className="flex items-center space-x-2">
              {darkMode ? <Moon className="w-5 h-5" /> : <Sun className="w-5 h-5" />}
              <span>界面设置</span>
            </CardTitle>
            <CardDescription>
              自定义界面显示偏好
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-6">
            <div className="flex items-center justify-between p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
              <div className="flex items-center space-x-3">
                <div className={`w-8 h-8 rounded-lg flex items-center justify-center ${darkMode ? 'bg-gray-700' : 'bg-gray-200'}`}>
                  {darkMode ? <Moon className="w-4 h-4" /> : <Sun className="w-4 h-4" />}
                </div>
                <div>
                  <h3 className="text-sm font-medium">深色模式</h3>
                  <p className="text-xs text-gray-500 dark:text-gray-400">
                    {darkMode ? '当前使用深色主题' : '当前使用浅色主题'}
                  </p>
                </div>
              </div>
              <Button 
                onClick={handleToggleDarkMode}
                variant="outline"
                size="sm"
              >
                {darkMode ? '切换到浅色' : '切换到深色'}
              </Button>
            </div>

            <div className="space-y-3">
              <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300">主题预览</h4>
              <div className="grid grid-cols-2 gap-3">
                <div className="p-3 bg-white border border-gray-200 rounded-lg">
                  <div className="w-full h-2 bg-gray-200 rounded mb-2"></div>
                  <div className="w-3/4 h-2 bg-gray-200 rounded mb-2"></div>
                  <div className="w-1/2 h-2 bg-gray-200 rounded"></div>
                  <p className="text-xs text-gray-500 mt-2">浅色主题</p>
                </div>
                <div className="p-3 bg-gray-800 border border-gray-700 rounded-lg">
                  <div className="w-full h-2 bg-gray-600 rounded mb-2"></div>
                  <div className="w-3/4 h-2 bg-gray-600 rounded mb-2"></div>
                  <div className="w-1/2 h-2 bg-gray-600 rounded"></div>
                  <p className="text-xs text-gray-400 mt-2">深色主题</p>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* System Configuration */}
        <div className="lg:col-span-2">
          <Card className="border-0 shadow-sm">
            <CardHeader>
              <CardTitle className="flex items-center space-x-2">
                <Sliders className="w-5 h-5" />
                <span>系统配置</span>
              </CardTitle>
              <CardDescription>
                配置系统运行参数和功能开关
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
              {configLoading ? (
                <div className="flex items-center justify-center py-12">
                  <RefreshCw className="w-6 h-6 animate-spin text-gray-400" />
                  <span className="ml-2 text-gray-500">加载配置中...</span>
                </div>
              ) : configError ? (
                <div className="flex items-center justify-center py-12">
                  <AlertCircle className="w-6 h-6 text-red-500" />
                  <span className="ml-2 text-red-600">{configError}</span>
                </div>
              ) : editingConfig && (
                <>
                  {/* Session Management */}
                  <div className="space-y-4">
                    <div className="flex items-center space-x-2">
                      <Activity className="w-5 h-5 text-blue-500" />
                      <h3 className="text-lg font-medium">Session管理</h3>
                      <Badge variant={editingConfig.sessionManagerEnabled ? "default" : "secondary"}>
                        {editingConfig.sessionManagerEnabled ? '已启用' : '已禁用'}
                      </Badge>
                    </div>
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                      <div className="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-800 rounded-lg">
                        <Label htmlFor="sessionManagerEnabled" className="text-sm font-medium">
                          启用Session管理
                        </Label>
                        <Switch
                          id="sessionManagerEnabled"
                          checked={editingConfig.sessionManagerEnabled}
                          onCheckedChange={(checked) =>
                            setEditingConfig({ ...editingConfig, sessionManagerEnabled: checked })
                          }
                        />
                      </div>
                      <div className="space-y-2">
                        <Label htmlFor="totalSessions" className="text-sm font-medium">
                          总Session数量
                        </Label>
                        <Input
                          id="totalSessions"
                          type="number"
                          min="1"
                          max="100"
                          value={editingConfig.totalSessions}
                          onChange={(e) =>
                            setEditingConfig({
                              ...editingConfig,
                              totalSessions: parseInt(e.target.value) || 1,
                            })
                          }
                        />
                      </div>
                      <div className="space-y-2">
                        <Label htmlFor="retryCount" className="text-sm font-medium">
                          重试次数
                        </Label>
                        <Input
                          id="retryCount"
                          type="number"
                          min="0"
                          max="10"
                          value={editingConfig.retryCount}
                          onChange={(e) =>
                            setEditingConfig({
                              ...editingConfig,
                              retryCount: parseInt(e.target.value) || 0,
                            })
                          }
                        />
                      </div>
                      <div className="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-800 rounded-lg">
                        <Label htmlFor="chatDelete" className="text-sm font-medium">
                          自动删除聊天记录
                        </Label>
                        <Switch
                          id="chatDelete"
                          checked={editingConfig.chatDelete}
                          onCheckedChange={(checked) =>
                            setEditingConfig({ ...editingConfig, chatDelete: checked })
                          }
                        />
                      </div>
                    </div>
                  </div>

                  {/* Chat Configuration */}
                  <div className="space-y-4">
                    <div className="flex items-center space-x-2">
                      <Database className="w-5 h-5 text-green-500" />
                      <h3 className="text-lg font-medium">聊天配置</h3>
                    </div>
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                      <div className="space-y-2">
                        <Label htmlFor="maxChatHistoryLength" className="text-sm font-medium">
                          最大聊天历史长度
                        </Label>
                        <Input
                          id="maxChatHistoryLength"
                          type="number"
                          min="1"
                          max="1000"
                          value={editingConfig.maxChatHistoryLength}
                          onChange={(e) =>
                            setEditingConfig({
                              ...editingConfig,
                              maxChatHistoryLength: parseInt(e.target.value) || 100,
                            })
                          }
                        />
                      </div>
                      <div className="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-800 rounded-lg">
                        <Label htmlFor="noRolePrefix" className="text-sm font-medium">
                          禁用角色前缀
                        </Label>
                        <Switch
                          id="noRolePrefix"
                          checked={editingConfig.noRolePrefix}
                          onCheckedChange={(checked) =>
                            setEditingConfig({ ...editingConfig, noRolePrefix: checked })
                          }
                        />
                      </div>
                      <div className="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-800 rounded-lg">
                        <Label htmlFor="promptDisableArtifacts" className="text-sm font-medium">
                          禁用Artifacts
                        </Label>
                        <Switch
                          id="promptDisableArtifacts"
                          checked={editingConfig.promptDisableArtifacts}
                          onCheckedChange={(checked) =>
                            setEditingConfig({
                              ...editingConfig,
                              promptDisableArtifacts: checked,
                            })
                          }
                        />
                      </div>
                    </div>
                  </div>

                  {/* API Configuration */}
                  <div className="space-y-4">
                    <div className="flex items-center space-x-2">
                      <Globe className="w-5 h-5 text-purple-500" />
                      <h3 className="text-lg font-medium">API配置</h3>
                    </div>
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                      <div className="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-800 rounded-lg">
                        <Label htmlFor="enableMirrorApi" className="text-sm font-medium">
                          启用镜像API
                        </Label>
                        <Switch
                          id="enableMirrorApi"
                          checked={editingConfig.enableMirrorApi}
                          onCheckedChange={(checked) =>
                            setEditingConfig({ ...editingConfig, enableMirrorApi: checked })
                          }
                        />
                      </div>
                      {editingConfig.enableMirrorApi && (
                        <div className="space-y-2">
                          <Label htmlFor="mirrorApiPrefix" className="text-sm font-medium">
                            镜像API前缀
                          </Label>
                          <Input
                            id="mirrorApiPrefix"
                            value={editingConfig.mirrorApiPrefix}
                            onChange={(e) =>
                              setEditingConfig({
                                ...editingConfig,
                                mirrorApiPrefix: e.target.value,
                              })
                            }
                            placeholder="/v1"
                          />
                        </div>
                      )}
                    </div>
                  </div>

                  {/* Session Manager Advanced Configuration */}
                  {editingConfig.sessionManagerEnabled && editingConfig.sessionManager && (
                    <div className="space-y-4">
                      <div className="flex items-center space-x-2">
                        <Shield className="w-5 h-5 text-orange-500" />
                        <h3 className="text-lg font-medium">Session管理器详细配置</h3>
                      </div>
                      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                        <div className="space-y-2">
                          <Label htmlFor="scheduleStrategy" className="text-sm font-medium">
                            调度策略
                          </Label>
                          <Input
                            id="scheduleStrategy"
                            value={editingConfig.sessionManager.scheduleStrategy}
                            onChange={(e) =>
                              setEditingConfig({
                                ...editingConfig,
                                sessionManager: {
                                  ...editingConfig.sessionManager!,
                                  scheduleStrategy: e.target.value,
                                },
                              })
                            }
                          />
                        </div>
                        <div className="space-y-2">
                          <Label htmlFor="maxRetryAttempts" className="text-sm font-medium">
                            最大重试次数
                          </Label>
                          <Input
                            id="maxRetryAttempts"
                            type="number"
                            min="1"
                            max="10"
                            value={editingConfig.sessionManager.maxRetryAttempts}
                            onChange={(e) =>
                              setEditingConfig({
                                ...editingConfig,
                                sessionManager: {
                                  ...editingConfig.sessionManager!,
                                  maxRetryAttempts: parseInt(e.target.value) || 3,
                                },
                              })
                            }
                          />
                        </div>
                        <div className="space-y-2">
                          <Label htmlFor="healthCheckInterval" className="text-sm font-medium">
                            健康检查间隔(秒)
                          </Label>
                          <Input
                            id="healthCheckInterval"
                            type="number"
                            min="1"
                            max="300"
                            value={editingConfig.sessionManager.healthCheckInterval}
                            onChange={(e) =>
                              setEditingConfig({
                                ...editingConfig,
                                sessionManager: {
                                  ...editingConfig.sessionManager!,
                                  healthCheckInterval: parseInt(e.target.value) || 30,
                                },
                              })
                            }
                          />
                        </div>
                        <div className="space-y-2">
                          <Label htmlFor="minHealthScore" className="text-sm font-medium">
                            最小健康分数
                          </Label>
                          <Input
                            id="minHealthScore"
                            type="number"
                            min="0"
                            max="1"
                            step="0.1"
                            value={editingConfig.sessionManager.minHealthScore}
                            onChange={(e) =>
                              setEditingConfig({
                                ...editingConfig,
                                sessionManager: {
                                  ...editingConfig.sessionManager!,
                                  minHealthScore: parseFloat(e.target.value) || 0.5,
                                },
                              })
                            }
                          />
                        </div>
                        <div className="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-800 rounded-lg">
                          <Label htmlFor="circuitBreakerEnabled" className="text-sm font-medium">
                            启用熔断器
                          </Label>
                          <Switch
                            id="circuitBreakerEnabled"
                            checked={editingConfig.sessionManager.circuitBreakerEnabled}
                            onCheckedChange={(checked) =>
                              setEditingConfig({
                                ...editingConfig,
                                sessionManager: {
                                  ...editingConfig.sessionManager!,
                                  circuitBreakerEnabled: checked,
                                },
                              })
                            }
                          />
                        </div>
                      </div>
                    </div>
                  )}

                  {/* Action Buttons */}
                  <div className="flex items-center justify-between pt-6 border-t border-gray-200 dark:border-gray-700">
                    <div className="flex space-x-3">
                      <Button
                        onClick={handleSaveConfig}
                        disabled={saveConfigStatus === 'saving'}
                      >
                        <Save className="w-4 h-4 mr-2" />
                        {saveConfigStatus === 'saving' ? '保存中...' : '保存配置'}
                      </Button>
                      <Button
                        variant="outline"
                        onClick={handleResetConfig}
                        disabled={saveConfigStatus === 'saving'}
                      >
                        <RotateCcw className="w-4 h-4 mr-2" />
                        重置
                      </Button>
                    </div>
                    {saveConfigStatus === 'success' && (
                      <div className="flex items-center space-x-2 text-sm text-green-600">
                        <Check className="w-4 h-4" />
                        <span>配置保存成功</span>
                      </div>
                    )}
                    {saveConfigStatus === 'error' && (
                      <div className="flex items-center space-x-2 text-sm text-red-600">
                        <AlertCircle className="w-4 h-4" />
                        <span>保存失败，请重试</span>
                      </div>
                    )}
                  </div>
                </>
              )}
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
};

export default SettingsPage;