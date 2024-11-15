// src/components/VerifyAccount.js
import React, { useEffect, useState } from 'react';
import { useLocation } from 'react-router-dom';
import './VerifyAccount.css'; // Create a CSS file for custom styling

const VerifyAccount = () => {
  const [message, setMessage] = useState('');
  const [loading, setLoading] = useState(true); // Loading state to show while verifying
  const location = useLocation();

  useEffect(() => {
    // Extract the token from the query parameter
    const searchParams = new URLSearchParams(location.search);
    const token = searchParams.get('token');

    // Mock verification process with loading state
    setTimeout(() => {
      if (token === 'mock-valid-token') {
        setMessage('Your account has been successfully verified!');
      } else {
        setMessage('The verification link is invalid or has expired. Please try again.');
      }
      setLoading(false);
    }, 1500); // Simulate a network delay
  }, [location]);

  return (
    <div className="verification-container">
      <div className="verification-box">
        {loading ? (
          <div className="loading-message">
            <span>Verifying your account...</span>
            <div className="loading-spinner"></div>
          </div>
        ) : (
          <p className={`verification-message ${message.includes('successfully') ? 'success' : 'error'}`}>
            {message}
          </p>
        )}
      </div>
    </div>
  );
};

export default VerifyAccount;
