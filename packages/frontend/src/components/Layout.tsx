import { Workflow, Key, LayoutTemplate } from 'lucide-react';
import { Link, Outlet, useLocation } from 'react-router-dom';

export default function Layout() {
  const location = useLocation();

  return (
    <div className="min-h-screen bg-gray-50">
      <nav className="bg-white border-b border-gray-200">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between h-16">
            <div className="flex">
              <Link to="/" className="flex items-center px-2 py-2 text-gray-900 font-bold text-xl">
                <Workflow className="mr-2" />
                Workflow Orchestration
              </Link>
              <div className="ml-6 flex space-x-4">
                <Link
                  to="/workflows"
                  className={`inline-flex items-center px-4 py-2 border-b-2 text-sm font-medium ${location.pathname.startsWith('/workflows')
                    ? 'border-blue-500 text-gray-900'
                    : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                    }`}
                >
                  <Workflow className="w-4 h-4 mr-2" />
                  Workflows
                </Link>
                <Link
                  to="/templates"
                  className={`inline-flex items-center px-4 py-2 border-b-2 text-sm font-medium ${location.pathname.startsWith('/templates')
                    ? 'border-blue-500 text-gray-900'
                    : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                    }`}
                >
                  <LayoutTemplate className="w-4 h-4 mr-2" />
                  Templates
                </Link>
                <Link
                  to="/credentials"
                  className={`inline-flex items-center px-4 py-2 border-b-2 text-sm font-medium ${location.pathname.startsWith('/credentials')
                    ? 'border-blue-500 text-gray-900'
                    : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                    }`}
                >
                  <Key className="w-4 h-4 mr-2" />
                  Credentials
                </Link>
              </div>
            </div>
          </div>
        </div>
      </nav>
      <main>
        <Outlet />
      </main>
    </div>
  );
}
