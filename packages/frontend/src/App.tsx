import { Routes, Route, Navigate } from 'react-router-dom';
import Layout from './components/Layout';
import WorkflowList from './pages/WorkflowList';
import WorkflowEditor from './pages/WorkflowEditor';
import ChatInterface from './pages/ChatInterface';
import CredentialsPage from './pages/CredentialsPage';
import TemplatesPage from './pages/TemplatesPage';

function App() {
  return (
    <Routes>
      <Route path="/" element={<Layout />}>
        <Route index element={<Navigate to="/workflows" replace />} />
        <Route path="workflows" element={<WorkflowList />} />
        <Route path="workflows/:id" element={<WorkflowEditor />} />
        <Route path="chat/:workflowId" element={<ChatInterface />} />
        <Route path="credentials" element={<CredentialsPage />} />
        <Route path="templates" element={<TemplatesPage />} />
      </Route>
    </Routes>
  );
}

export default App;
