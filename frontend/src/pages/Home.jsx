import React, { useState, useEffect } from 'react';
import { Copy, Link2, TrendingUp, Github, ExternalLink, LogIn, LogOut } from 'lucide-react';

const API_BASE_URL = 'http://localhost:8080';

// Cookie utility functions
const getCookie = (name) => {
  const value = `; ${document.cookie}`;
  const parts = value.split(`; ${name}=`);
  if (parts.length === 2) return parts.pop().split(';').shift();
  return null;
};

export default function ShortyApp() {
  const [url, setUrl] = useState('');
  const [shortUrl, setShortUrl] = useState('');
  const [loading, setLoading] = useState(false);
  const [links, setLinks] = useState([]);
  const [copied, setCopied] = useState(false);
  const [error, setError] = useState('');
  const [userId, setUserId] = useState(null);
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [loadingLinks, setLoadingLinks] = useState(false);

  // Check authentication on mount
  useEffect(() => {
    const userIdFromCookie = getCookie('userId');
    if (userIdFromCookie) {
      setUserId(userIdFromCookie);
      setIsAuthenticated(true);
    }
  }, []);

  // Fetch links when authenticated
  useEffect(() => {
    if (isAuthenticated && userId) {
      fetchLinks();
    }
  }, [isAuthenticated, userId]);

  const fetchLinks = async () => {
    if (!isAuthenticated || !userId) return;
    
    setLoadingLinks(true);
    try {
      const response = await fetch(`${API_BASE_URL}/api/urls/stats?user_id=${userId}`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
      });
      if (response.ok) {
        const data = await response.json();
        // Transform API response to match component structure
        const transformedLinks = Array.isArray(data) ? data.map((item, index) => ({
          id: index + Date.now(),
          shortCode: item.short_url.split('/').pop(),
          shortUrl: item.short_url,
          originalUrl: item.original_url,
          clicks: item.clicks || 0,
          createdAt: new Date().toISOString(),
          preview: null
        })) : [];
        setLinks(transformedLinks);
      } else {
        setLinks([]);
      }
    } catch (err) {
      console.error('Failed to fetch links:', err);
      setLinks([]);
    } finally {
      setLoadingLinks(false);
    }
  };

  const handleShorten = async () => {
    if (!isAuthenticated) {
      setError('Please login to shorten URLs');
      return;
    }

    if (!url) {
      setError('Please enter a URL');
      return;
    }

    setLoading(true);
    setError('');
    
    try {
      const response = await fetch(`${API_BASE_URL}/api/urls`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
        body: JSON.stringify({
          original_url: url
        })
      });

      if (!response.ok) {
        const errorText = await response.text();
        try {
          const errorJson = JSON.parse(errorText);
          throw new Error(errorJson.error || 'Failed to shorten URL');
        } catch (e) {
          throw new Error(errorText || 'Failed to shorten URL');
        }
      }

      const data = await response.json();
      
      const newShortUrl = data.short_url;
      setShortUrl(newShortUrl);
      
      const newLink = {
        id: data.id || Date.now(),
        shortCode: data.short_code || data.shortCode,
        shortUrl: newShortUrl,
        originalUrl: url,
        clicks: data.clicks || 0,
        createdAt: data.created_at || new Date().toISOString(),
        preview: data.preview || null
      };
      
      setLinks(prevLinks => [newLink, ...prevLinks]);
      setUrl('');
      
      // Refresh links list from API
      fetchLinks();
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const copyToClipboard = (text) => {
    navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const handleLogin = () => {
    // Redirect to login page
    window.location.href = '/login';
  };

  const handleLogout = () => {
    // Clear cookies and localStorage
    document.cookie = 'userId=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;';
    localStorage.removeItem('token');
    setUserId(null);
    setIsAuthenticated(false);
    setLinks([]);
    setShortUrl('');
  };

  const formatDate = (dateString) => {
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
  };

  return (
    <div className="min-vh-100" style={{ backgroundColor: '#f8f9fa' }}>
      {/* Header */}
      <nav className="navbar navbar-light bg-white border-bottom">
        <div className="container-fluid px-4">
          <a className="navbar-brand d-flex align-items-center" href="#">
            <div className="d-flex align-items-center justify-content-center bg-dark text-white rounded me-2" 
                 style={{ width: '32px', height: '32px', fontWeight: '700' }}>
              S
            </div>
            <span className="fw-bold fs-5">Shorty</span>
          </a>
          <div className="d-flex align-items-center gap-3">
            <a href="https://github.com" className="text-decoration-none text-muted d-flex align-items-center">
              <Github size={18} className="me-2" />
              GitHub
            </a>
            {isAuthenticated ? (
              <button 
                onClick={handleLogout}
                className="btn btn-outline-dark btn-sm d-flex align-items-center"
              >
                <LogOut size={16} className="me-2" />
                Logout
              </button>
            ) : (
              <button 
                onClick={handleLogin}
                className="btn btn-dark btn-sm d-flex align-items-center"
              >
                <LogIn size={16} className="me-2" />
                Login
              </button>
            )}
          </div>
        </div>
      </nav>

      {/* Main Content */}
      <div className="container" style={{ maxWidth: '1200px', padding: '3rem 1rem' }}>
        {/* Hero Section */}
        <div className="text-center mb-5">
          <h1 className="display-4 fw-bold mb-3">Simplify your links, amplify your reach.</h1>
          <p className="lead text-muted">Create short, powerful links that are easy to share and track.</p>
        </div>

        {/* Login Warning */}
        {!isAuthenticated && (
          <div className="alert alert-warning mb-4" role="alert">
            <div className="d-flex align-items-center">
              <LogIn size={20} className="me-3" />
              <div>
                <strong>Authentication Required</strong>
                <p className="mb-0 mt-1">Please login to shorten URLs and view your link statistics.</p>
              </div>
            </div>
          </div>
        )}

        {/* Shorten Box */}
        <div className="card shadow-sm mb-5 border-0">
          <div className="card-body p-4">
            <div className="row g-3">
              <div className="col-md-9">
                <div className="input-group input-group-lg">
                  <span className="input-group-text bg-white border-end-0">
                    <Link2 size={20} className="text-muted" />
                  </span>
                  <input
                    type="url"
                    className="form-control border-start-0 ps-0"
                    placeholder="Paste your long link here..."
                    value={url}
                    onChange={(e) => setUrl(e.target.value)}
                    onKeyPress={(e) => e.key === 'Enter' && handleShorten()}
                    disabled={!isAuthenticated}
                  />
                </div>
              </div>
              <div className="col-md-3">
                <button 
                  onClick={handleShorten}
                  className="btn btn-dark btn-lg w-100"
                  disabled={loading || !isAuthenticated}
                >
                  {loading ? (
                    <>
                      <span className="spinner-border spinner-border-sm me-2" role="status" aria-hidden="true"></span>
                      Shortening...
                    </>
                  ) : 'Shorten Now'}
                </button>
              </div>
            </div>

            {/* Error */}
            {error && (
              <div className="alert alert-danger mt-3 mb-0" role="alert">
                {error}
              </div>
            )}

            {/* Result */}
            {shortUrl && (
              <div className="alert alert-success mt-4 mb-0" role="alert">
                <div className="d-flex justify-content-between align-items-center mb-2">
                  <small className="fw-semibold">YOUR SHORTENED URL</small>
                </div>
                <div className="d-flex align-items-center bg-white rounded p-3">
                  <div className="flex-grow-1 me-3">
                    <a href={shortUrl} className="text-primary fw-semibold text-decoration-none fs-5" target="_blank" rel="noopener noreferrer">
                      {shortUrl}
                    </a>
                  </div>
                  <button 
                    className="btn btn-outline-primary btn-sm"
                    onClick={() => copyToClipboard(shortUrl)}
                  >
                    <Copy size={16} className="me-1" />
                    {copied ? 'Copied!' : 'Copy'}
                  </button>
                </div>
              </div>
            )}
          </div>
        </div>

        {/* Recent Links - Only show when authenticated */}
        {isAuthenticated && (
          <div className="card shadow-sm border-0">
            <div className="card-body p-4">
              <div className="d-flex justify-content-between align-items-center mb-4">
                <h2 className="h4 fw-bold mb-0">Your Links</h2>
                <span className="text-muted small">Manage and track your shortened URLs</span>
              </div>

              {links.length === 0 ? (
                <div className="text-center py-5">
                  <div className="text-muted">
                    <Link2 size={48} className="mb-3 opacity-25" />
                    <p className="mb-0">No links generated yet. Start by shortening one above!</p>
                  </div>
                </div>
              ) : (
                <div className="list-group list-group-flush">
                  {links.map((link) => (
                    <div key={link.id} className="list-group-item px-0 py-3 border-bottom">
                      <div className="row align-items-start">
                        {/* Preview Image */}
                        {link.preview && (
                          <div className="col-md-2 mb-3 mb-md-0">
                            <img 
                              src={link.preview} 
                              alt="Link preview"
                              className="img-fluid rounded"
                              style={{ width: '100%', height: '80px', objectFit: 'cover' }}
                              onError={(e) => {
                                e.target.style.display = 'none';
                              }}
                            />
                          </div>
                        )}
                        
                        {/* Link Info */}
                        <div className={link.preview ? "col-md-6" : "col-md-7"}>
                          <div className="mb-2">
                            <a href={link.shortUrl} className="text-decoration-none fw-semibold text-dark d-inline-flex align-items-center" target="_blank" rel="noopener noreferrer">
                              {link.shortUrl}
                              <ExternalLink size={14} className="ms-2 text-muted" />
                            </a>
                          </div>
                          <div className="small text-muted text-truncate" style={{ maxWidth: '400px' }}>
                            {link.originalUrl}
                          </div>
                        </div>
                        
                        {/* Stats */}
                        <div className="col-md-2">
                          <div className="d-flex align-items-center text-muted small">
                            <TrendingUp size={16} className="me-2" />
                            <span className="fw-semibold">{link.clicks}</span>
                            <span className="ms-1">clicks</span>
                          </div>
                          <div className="text-muted small mt-1">
                            {formatDate(link.createdAt)}
                          </div>
                        </div>
                        
                        {/* Actions */}
                        <div className="col-md-2 text-end">
                          <button 
                            className="btn btn-outline-secondary btn-sm"
                            onClick={() => copyToClipboard(link.shortUrl)}
                          >
                            <Copy size={14} className="me-1" />
                            Copy
                          </button>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        )}
      </div>

      {/* Bootstrap CSS */}
      <link 
        href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" 
        rel="stylesheet" 
        integrity="sha384-9ndCyUaIbzAi2FUVXJi0CjmCapSmO7SnpJef0486qhLnuZ2cdeRhO02iuK6FUUVM" 
        crossOrigin="anonymous"
      />
    </div>
  );
}