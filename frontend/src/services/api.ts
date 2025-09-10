import axios, { AxiosInstance } from 'axios';
import { Model, SessionHealth, SystemStats } from '../types';

class ApiService {
  private api: AxiosInstance;

  constructor() {
    const baseURL = import.meta.env.VITE_API_BASE_URL || '';
    this.api = axios.create({
      baseURL,
      timeout: 30000,
    });

    // 请求拦截器
    this.api.interceptors.request.use(
      (config) => {
        const apiKey = localStorage.getItem('apiKey') || 'test-api-key-123';
        if (apiKey) {
          config.headers.Authorization = `Bearer ${apiKey}`;
        }
        return config;
      },
      (error) => Promise.reject(error)
    );

    // 响应拦截器
    this.api.interceptors.response.use(
      (response) => response,
      (error) => {
        if (error.response?.status === 401) {
          localStorage.removeItem('apiKey');
          window.location.href = '/settings';
        }
        return Promise.reject(error);
      }
    );
  }

  // 获取模型列表
  async getModels(): Promise<{ data: Model[] }> {
    const response = await this.api.get('/v1/models');
    return response.data;
  }

  // 聊天完成
  async chatCompletion(messages: any[], model: string, stream: boolean = true) {
    const response = await this.api.post('/v1/chat/completions', {
      model,
      messages,
      stream,
    });
    return response.data;
  }

  // 流式聊天
  async streamChat(
    messages: any[],
    model: string,
    onMessage: (message: string) => void,
    onThinking?: (thinking: string) => void
  ) {
    const base = import.meta.env.VITE_API_BASE_URL || '';
    const response = await fetch(`${base}/v1/chat/completions`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('apiKey') || 'test-api-key-123'}`,
      },
      body: JSON.stringify({
        model,
        messages,
        stream: true,
      }),
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    const reader = response.body?.getReader();
    const decoder = new TextDecoder();

    if (!reader) {
      throw new Error('Stream reader not available');
    }

    let buffer = '';
    while (true) {
      const { done, value } = await reader.read();
      if (done) break;

      buffer += decoder.decode(value, { stream: true });
      const lines = buffer.split('\n');
      buffer = lines.pop() || '';

      for (const line of lines) {
        if (line.startsWith('data: ')) {
          const data = line.slice(6);
          if (data === '[DONE]') {
            return;
          }
          try {
            const parsed = JSON.parse(data);
            const content = parsed.choices?.[0]?.delta?.content || '';
            const thinking = parsed.choices?.[0]?.delta?.thinking || '';
            
            if (thinking && onThinking) {
              onThinking(thinking);
            }
            if (content) {
              onMessage(content);
            }
          } catch (e) {
            // 忽略解析错误
          }
        }
      }
    }
  }

  // 获取会话健康状态
  async getSessionHealth(): Promise<SessionHealth[]> {
    const response = await this.api.get('/admin/sessions');
    return response.data;
  }

  // 获取系统统计
  async getSystemStats(): Promise<SystemStats> {
    const response = await this.api.get('/admin/stats');
    return response.data;
  }

  // 获取配置
  async getConfig() {
    const response = await this.api.get('/admin/config');
    return response.data;
  }

  // 健康检查
  async healthCheck() {
    const response = await this.api.get('/health');
    return response.data;
  }
}

export const apiService = new ApiService();
