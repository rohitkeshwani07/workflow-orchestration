import { useState, useEffect } from 'react';
import { Plus, Trash2, Edit, Eye, EyeOff, Key } from 'lucide-react';

interface Credential {
  id: string;
  name: string;
  type: string;
  description?: string;
  value?: string;
  metadata?: Record<string, any>;
  createdAt: string;
  updatedAt: string;
}

export default function CredentialsPage() {
  const [credentials, setCredentials] = useState<Credential[]>([]);
  const [loading, setLoading] = useState(true);
  const [showModal, setShowModal] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [showValue, setShowValue] = useState<Record<string, boolean>>({});
  const [formData, setFormData] = useState({
    name: '',
    type: 'api_key' as 'api_key' | 'oauth' | 'basic_auth' | 'bearer_token' | 'custom',
    description: '',
    value: '',
  });

  const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:3001';

  useEffect(() => {
    fetchCredentials();
  }, []);

  const fetchCredentials = async () => {
    try {
      const response = await fetch(`${API_BASE_URL}/api/credentials`);
      const result = await response.json();
      if (result.success) {
        setCredentials(result.data);
      }
    } catch (error) {
      console.error('Failed to fetch credentials:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleCreate = async () => {
    try {
      const response = await fetch(`${API_BASE_URL}/api/credentials`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(formData),
      });
      const result = await response.json();
      if (result.success) {
        await fetchCredentials();
        closeModal();
      }
    } catch (error) {
      console.error('Failed to create credential:', error);
    }
  };

  const handleUpdate = async () => {
    if (!editingId) return;
    try {
      const response = await fetch(`${API_BASE_URL}/api/credentials/${editingId}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(formData),
      });
      const result = await response.json();
      if (result.success) {
        await fetchCredentials();
        closeModal();
      }
    } catch (error) {
      console.error('Failed to update credential:', error);
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to delete this credential?')) return;
    try {
      const response = await fetch(`${API_BASE_URL}/api/credentials/${id}`, {
        method: 'DELETE',
      });
      const result = await response.json();
      if (result.success) {
        await fetchCredentials();
      }
    } catch (error) {
      console.error('Failed to delete credential:', error);
    }
  };

  const openModal = (credential?: Credential) => {
    if (credential) {
      setEditingId(credential.id);
      setFormData({
        name: credential.name,
        type: credential.type as any,
        description: credential.description || '',
        value: '', // Don't pre-fill value for security
      });
    } else {
      setEditingId(null);
      setFormData({
        name: '',
        type: 'api_key',
        description: '',
        value: '',
      });
    }
    setShowModal(true);
  };

  const closeModal = () => {
    setShowModal(false);
    setEditingId(null);
    setFormData({
      name: '',
      type: 'api_key',
      description: '',
      value: '',
    });
  };

  const toggleShowValue = async (id: string) => {
    if (showValue[id]) {
      // Hide value
      setShowValue({ ...showValue, [id]: false });
    } else {
      // Fetch and show value
      try {
        const response = await fetch(`${API_BASE_URL}/api/credentials/${id}`);
        const result = await response.json();
        if (result.success) {
          setCredentials(credentials.map(c =>
            c.id === id ? { ...c, value: result.data.value } : c
          ));
          setShowValue({ ...showValue, [id]: true });
        }
      } catch (error) {
        console.error('Failed to fetch credential value:', error);
      }
    }
  };

  const getTypeLabel = (type: string) => {
    const labels: Record<string, string> = {
      api_key: 'API Key',
      oauth: 'OAuth',
      basic_auth: 'Basic Auth',
      bearer_token: 'Bearer Token',
      custom: 'Custom',
    };
    return labels[type] || type;
  };

  const getTypeColor = (type: string) => {
    const colors: Record<string, string> = {
      api_key: 'bg-blue-100 text-blue-800',
      oauth: 'bg-green-100 text-green-800',
      basic_auth: 'bg-yellow-100 text-yellow-800',
      bearer_token: 'bg-purple-100 text-purple-800',
      custom: 'bg-gray-100 text-gray-800',
    };
    return colors[type] || 'bg-gray-100 text-gray-800';
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-screen">
        <div className="text-lg">Loading credentials...</div>
      </div>
    );
  }

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-3xl font-bold text-gray-900">Credentials</h1>
        <button
          onClick={() => openModal()}
          className="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700"
        >
          <Plus className="w-4 h-4 mr-2" />
          New Credential
        </button>
      </div>

      {credentials.length === 0 ? (
        <div className="text-center py-12 bg-white rounded-lg border border-gray-200">
          <Key className="mx-auto h-12 w-12 text-gray-400" />
          <h3 className="mt-2 text-sm font-medium text-gray-900">No credentials</h3>
          <p className="mt-1 text-sm text-gray-500">Get started by creating a new credential.</p>
          <div className="mt-6">
            <button
              onClick={() => openModal()}
              className="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700"
            >
              <Plus className="w-4 h-4 mr-2" />
              New Credential
            </button>
          </div>
        </div>
      ) : (
        <div className="bg-white shadow overflow-hidden sm:rounded-md">
          <ul className="divide-y divide-gray-200">
            {credentials.map((credential) => (
              <li key={credential.id}>
                <div className="px-6 py-4">
                  <div className="flex items-center justify-between">
                    <div className="flex-1">
                      <div className="flex items-center">
                        <h3 className="text-lg font-medium text-gray-900">{credential.name}</h3>
                        <span className={`ml-3 px-2.5 py-0.5 rounded-full text-xs font-medium ${getTypeColor(credential.type)}`}>
                          {getTypeLabel(credential.type)}
                        </span>
                      </div>
                      {credential.description && (
                        <p className="mt-1 text-sm text-gray-500">{credential.description}</p>
                      )}
                      <div className="mt-2 flex items-center space-x-2">
                        <code className="px-2 py-1 bg-gray-100 rounded text-sm font-mono">
                          {showValue[credential.id] ? credential.value : '••••••••••••••••'}
                        </code>
                        <button
                          onClick={() => toggleShowValue(credential.id)}
                          className="text-gray-400 hover:text-gray-600"
                        >
                          {showValue[credential.id] ? (
                            <EyeOff className="w-4 h-4" />
                          ) : (
                            <Eye className="w-4 h-4" />
                          )}
                        </button>
                      </div>
                      <p className="mt-2 text-xs text-gray-400">
                        Created: {new Date(credential.createdAt).toLocaleDateString()}
                      </p>
                    </div>
                    <div className="flex space-x-2">
                      <button
                        onClick={() => openModal(credential)}
                        className="p-2 text-gray-400 hover:text-blue-600"
                      >
                        <Edit className="w-5 h-5" />
                      </button>
                      <button
                        onClick={() => handleDelete(credential.id)}
                        className="p-2 text-gray-400 hover:text-red-600"
                      >
                        <Trash2 className="w-5 h-5" />
                      </button>
                    </div>
                  </div>
                </div>
              </li>
            ))}
          </ul>
        </div>
      )}

      {/* Modal */}
      {showModal && (
        <div className="fixed inset-0 bg-gray-500 bg-opacity-75 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg shadow-xl max-w-md w-full mx-4">
            <div className="px-6 py-4 border-b border-gray-200">
              <h2 className="text-xl font-semibold text-gray-900">
                {editingId ? 'Edit Credential' : 'New Credential'}
              </h2>
            </div>
            <div className="px-6 py-4 space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Name</label>
                <input
                  type="text"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                  placeholder="My API Key"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Type</label>
                <select
                  value={formData.type}
                  onChange={(e) => setFormData({ ...formData, type: e.target.value as any })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                >
                  <option value="api_key">API Key</option>
                  <option value="oauth">OAuth</option>
                  <option value="basic_auth">Basic Auth</option>
                  <option value="bearer_token">Bearer Token</option>
                  <option value="custom">Custom</option>
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Description (optional)</label>
                <input
                  type="text"
                  value={formData.description}
                  onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                  placeholder="Description"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Value {editingId && '(leave empty to keep current)'}
                </label>
                <input
                  type="password"
                  value={formData.value}
                  onChange={(e) => setFormData({ ...formData, value: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                  placeholder="sk-..."
                />
              </div>
            </div>
            <div className="px-6 py-4 bg-gray-50 border-t border-gray-200 flex justify-end space-x-3">
              <button
                onClick={closeModal}
                className="px-4 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 hover:bg-gray-50"
              >
                Cancel
              </button>
              <button
                onClick={editingId ? handleUpdate : handleCreate}
                className="px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700"
              >
                {editingId ? 'Update' : 'Create'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
