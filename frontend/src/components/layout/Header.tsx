import React from 'react';
import { Menu, Bot, Settings } from 'lucide-react';
import { Link } from 'react-router-dom';
import { Button } from '@/components/ui/button';
import { useChatStore } from '@/stores/chat';

const Header: React.FC = () => {
  const { toggleSidebar } = useChatStore();

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
          <h1 className="text-lg font-semibold">
            Claude2Api+ 管理界面
          </h1>
        </div>
      </div>

      <div className="flex items-center space-x-4">
        <Link to="/settings">
          <Button variant="ghost" size="icon" aria-label="设置">
            <Settings className="h-4 w-4" />
          </Button>
        </Link>
      </div>
    </header>
  );
};

export default Header;
