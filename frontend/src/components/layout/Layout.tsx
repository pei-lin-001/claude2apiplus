import React from 'react';
// Note: We render children directly because App comp wraps <Routes> inside <Layout>
import Sidebar from '@/components/layout/Sidebar';
import Header from '@/components/layout/Header';
import { useChatStore } from '@/stores/chat';

interface LayoutProps {
  children?: React.ReactNode;
}

const Layout: React.FC<LayoutProps> = ({ children }) => {
  const { sidebarOpen } = useChatStore();

  return (
    <div className="flex h-screen bg-background">
      <Sidebar />
      <div className={`flex-1 flex flex-col ${sidebarOpen ? 'ml-64' : 'ml-0'} transition-all duration-300`}>
        <Header />
        <main className="flex-1 overflow-auto">{children}</main>
      </div>
    </div>
  );
};

export default Layout;
