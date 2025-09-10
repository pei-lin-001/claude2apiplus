import { useEffect } from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import { useChatStore } from '@/stores/chat';
import Layout from '@/components/layout/Layout';
import SessionManager from '@/pages/SessionManager';
import StatisticsPage from '@/pages/StatisticsPage';
import SettingsPage from '@/pages/SettingsPage';

function App() {
  const { darkMode } = useChatStore();

  useEffect(() => {
    // 应用深色模式
    if (darkMode) {
      document.documentElement.classList.add('dark');
    } else {
      document.documentElement.classList.remove('dark');
    }
  }, [darkMode]);

  return (
    <Layout>
      <Routes>
        <Route path="/" element={<Navigate to="/sessions" replace />} />
        <Route path="/sessions" element={<SessionManager />} />
        <Route path="/statistics" element={<StatisticsPage />} />
        <Route path="/settings" element={<SettingsPage />} />
      </Routes>
    </Layout>
  );
}

export default App;
