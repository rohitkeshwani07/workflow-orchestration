import { useState, useEffect } from 'react';
import { LayoutTemplate, Plus, Tag } from 'lucide-react';
import { useNavigate } from 'react-router-dom';

interface WorkflowTemplate {
  id: string;
  name: string;
  description?: string;
  category: string;
  tags: string[];
  nodes: any[];
  edges: any[];
  thumbnail?: string;
  featured: boolean;
  createdAt: string;
  updatedAt: string;
}

export default function TemplatesPage() {
  const [templates, setTemplates] = useState<WorkflowTemplate[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedCategory, setSelectedCategory] = useState<string>('all');
  const navigate = useNavigate();

  const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:3001';

  useEffect(() => {
    fetchTemplates();
  }, [selectedCategory]);

  const fetchTemplates = async () => {
    try {
      const url = selectedCategory === 'all'
        ? `${API_BASE_URL}/api/templates`
        : `${API_BASE_URL}/api/templates?category=${encodeURIComponent(selectedCategory)}`;
      const response = await fetch(url);
      const result = await response.json();
      if (result.success) {
        setTemplates(result.data);
      }
    } catch (error) {
      console.error('Failed to fetch templates:', error);
    } finally {
      setLoading(false);
    }
  };

  const createWorkflowFromTemplate = async (templateId: string, templateName: string) => {
    try {
      const response = await fetch(`${API_BASE_URL}/api/templates/${templateId}/create-workflow`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          name: `${templateName} (Copy)`,
        }),
      });
      const result = await response.json();
      if (result.success) {
        navigate(`/workflows/${result.data.id}`);
      }
    } catch (error) {
      console.error('Failed to create workflow from template:', error);
    }
  };

  const categories = ['all', ...Array.from(new Set(templates.map(t => t.category)))];

  const getCategoryColor = (category: string) => {
    const colors: Record<string, string> = {
      'Customer Support': 'bg-blue-100 text-blue-800',
      'Data Processing': 'bg-green-100 text-green-800',
      'Content Moderation': 'bg-red-100 text-red-800',
      'Sales': 'bg-purple-100 text-purple-800',
      'Development': 'bg-yellow-100 text-yellow-800',
    };
    return colors[category] || 'bg-gray-100 text-gray-800';
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-screen">
        <div className="text-lg">Loading templates...</div>
      </div>
    );
  }

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <div className="mb-6">
        <h1 className="text-3xl font-bold text-gray-900">Workflow Templates</h1>
        <p className="mt-2 text-sm text-gray-600">
          Start with a pre-built workflow template and customize it to your needs
        </p>
      </div>

      {/* Category filter */}
      <div className="mb-6 flex items-center space-x-2 overflow-x-auto pb-2">
        {categories.map((category) => (
          <button
            key={category}
            onClick={() => setSelectedCategory(category)}
            className={`px-4 py-2 rounded-full text-sm font-medium whitespace-nowrap ${
              selectedCategory === category
                ? 'bg-blue-600 text-white'
                : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
            }`}
          >
            {category === 'all' ? 'All Templates' : category}
          </button>
        ))}
      </div>

      {templates.length === 0 ? (
        <div className="text-center py-12 bg-white rounded-lg border border-gray-200">
          <LayoutTemplate className="mx-auto h-12 w-12 text-gray-400" />
          <h3 className="mt-2 text-sm font-medium text-gray-900">No templates found</h3>
          <p className="mt-1 text-sm text-gray-500">
            {selectedCategory === 'all'
              ? 'No templates available yet.'
              : `No templates in the "${selectedCategory}" category.`}
          </p>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {templates.map((template) => (
            <div
              key={template.id}
              className="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden hover:shadow-md transition-shadow"
            >
              {/* Thumbnail */}
              <div className="h-40 bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center">
                <LayoutTemplate className="w-16 h-16 text-blue-400" />
              </div>

              {/* Content */}
              <div className="p-5">
                <div className="flex items-start justify-between mb-2">
                  <h3 className="text-lg font-semibold text-gray-900">{template.name}</h3>
                  {template.featured && (
                    <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-yellow-100 text-yellow-800">
                      Featured
                    </span>
                  )}
                </div>

                <p className="text-sm text-gray-600 mb-3 line-clamp-2">
                  {template.description || 'No description available'}
                </p>

                <div className="flex items-center justify-between mb-3">
                  <span className={`px-2.5 py-0.5 rounded-full text-xs font-medium ${getCategoryColor(template.category)}`}>
                    {template.category}
                  </span>
                  <span className="text-xs text-gray-500">
                    {template.nodes.length} nodes
                  </span>
                </div>

                {template.tags && template.tags.length > 0 && (
                  <div className="flex flex-wrap gap-1 mb-4">
                    {template.tags.slice(0, 3).map((tag, idx) => (
                      <span
                        key={idx}
                        className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-gray-100 text-gray-600"
                      >
                        <Tag className="w-3 h-3 mr-1" />
                        {tag}
                      </span>
                    ))}
                    {template.tags.length > 3 && (
                      <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-gray-100 text-gray-600">
                        +{template.tags.length - 3}
                      </span>
                    )}
                  </div>
                )}

                <button
                  onClick={() => createWorkflowFromTemplate(template.id, template.name)}
                  className="w-full inline-flex items-center justify-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700"
                >
                  <Plus className="w-4 h-4 mr-2" />
                  Use Template
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
