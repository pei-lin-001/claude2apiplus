import React from 'react';
import { Menu, Bot, Settings, LogOut } from 'lucide-react';
import { Link } from 'react-router-dom';
import { Button } from '@/components/ui/button';
import { useChatStore } from '@/stores/chat';
import { useAuthStore } from '@/stores/auth';
import { apiService } from '@/services/api';

const Header: React.FC = () => {
  const { toggleSidebar } = useChatStore();
  const { user, logout } = useAuthStore();

  const handleLogout = async () => {
    try {
      await apiService.adminLogout();
    } catch (error) {
      console.error('Logout failed:', error);
    } finally {
      logout();
      window.location.href = '/login';
    }
  };

  return (
    <header className="h-16 border-b border-border bg-card flex items-center justify-between px-4">
      <div className="flex items-center space-x-4">
        <Button
          variant="ghost"
          size="icon"
          onClick={toggleSidebar}
          className="lg:hidden"
        >
          <Menu className="h-5 w-5" />
        </Button>
        
        <div className="flex items-center space-x-2">
          <Bot className="h-6 w-6 text-primary" />
          <div>
            <h1 className="text-lg font-semibold">
              Claude2Api+ 管理界面
            </h1>
            {user && (
              <p className="text-xs text-muted-foreground">
                欢迎, {user.username}
              </p>
            )}
          </div>
        </div>
      </div>

      <div className="flex items-center space-x-4">
        <Link to="/settings">
          <Button variant="ghost" size="icon" aria-label="设置">
            <Settings className="h-4 w-4" />
          </Button>
        </Link>
        
        <Button
          variant="ghost"
          size="icon"
          onClick={handleLogout}
          aria-label="登出"
          title="登出"
        >
          <LogOut className="h-4 w-4" />
        </Button>
      </div>
    </header>
  );
};

export default Header;
