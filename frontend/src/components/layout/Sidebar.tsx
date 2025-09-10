import React from 'react';
import { Link, useLocation } from 'react-router-dom';
import { 
  Key, 
  Settings, 
  BarChart3, 
  X,
  Moon,
  Sun
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { useChatStore } from '@/stores/chat';

const Sidebar: React.FC = () => {
  const {
    sidebarOpen,
    toggleSidebar,
    toggleDarkMode,
    darkMode,
  } = useChatStore();

  const location = useLocation();

  return (
    <>
      {/* 移动端遮罩 */}
      {sidebarOpen && (
        <div
          className="fixed inset-0 bg-black bg-opacity-50 z-40 lg:hidden"
          onClick={toggleSidebar}
        />
      )}
      
      {/* 侧边栏 */}
      <div
        className={`fixed left-0 top-0 h-full w-64 bg-card border-r border-border z-50 transform transition-transform duration-300 ${
          sidebarOpen ? 'translate-x-0' : '-translate-x-full'
        } lg:translate-x-0`}
      >
        <div className="flex flex-col h-full">
          {/* 头部 */}
          <div className="flex items-center justify-between p-4 border-b border-border">
            <h1 className="text-xl font-semibold">Claude2Api+</h1>
            <Button
              variant="ghost"
              size="icon"
              onClick={toggleSidebar}
              className="lg:hidden"
            >
              <X className="h-4 w-4" />
            </Button>
          </div>

          {/* 导航菜单 */}
          <div className="flex-1 overflow-y-auto p-4">
            <nav className="space-y-1">
              <Link
                to="/sessions"
                className={`flex items-center px-3 py-2 rounded-lg text-sm transition-colors ${
                  location.pathname === '/sessions' ? 'bg-accent text-accent-foreground' : 'text-muted-foreground hover:text-foreground hover:bg-accent'
                }`}
              >
                <Key className="h-4 w-4 mr-3" />
                Session管理
              </Link>
              <Link
                to="/statistics"
                className={`flex items-center px-3 py-2 rounded-lg text-sm transition-colors ${
                  location.pathname === '/statistics' ? 'bg-accent text-accent-foreground' : 'text-muted-foreground hover:text-foreground hover:bg-accent'
                }`}
              >
                <BarChart3 className="h-4 w-4 mr-3" />
                状态统计
              </Link>
              <Link
                to="/settings"
                className={`flex items-center px-3 py-2 rounded-lg text-sm transition-colors ${
                  location.pathname === '/settings' ? 'bg-accent text-accent-foreground' : 'text-muted-foreground hover:text-foreground hover:bg-accent'
                }`}
              >
                <Settings className="h-4 w-4 mr-3" />
                设置
              </Link>
            </nav>
          </div>

          {/* 底部主题切换 */}
          <div className="border-t border-border p-4">
            <Button
              variant="ghost"
              size="sm"
              onClick={toggleDarkMode}
              className="w-full justify-start"
            >
              {darkMode ? (
                <>
                  <Sun className="h-4 w-4 mr-3" />
                  浅色模式
                </>
              ) : (
                <>
                  <Moon className="h-4 w-4 mr-3" />
                  深色模式
                </>
              )}
            </Button>
          </div>
        </div>
      </div>
    </>
  );
};

export default Sidebar;