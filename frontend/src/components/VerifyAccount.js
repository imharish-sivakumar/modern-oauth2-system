// src/components/VerifyAccount.js
import React, { useEffect, useState } from 'react';
import { useLocation } from 'react-router-dom';
import './VerifyAccount.css'; // Create a CSS file for custom styling
import loginService from "../services/loginService";
import encrypt from "../helper/encrypt";

const VerifyAccount = () => {
  const [message, setMessage] = useState('');
  const [loading, setLoading] = useState(true); // Loading state to show while verifying
  const location = useLocation();

  useEffect(() => {
    // Extract the token from the query parameter
    const searchParams = new URLSearchParams(location.search);
    const token = searchParams.get('token');

    if (token) {
      // Call the verifyAccount API
      loginService.verifyAccount(token)
        .then((response) => {
          if (response.status === 200) {
            setMessage('Your account has been successfully verified!');
          } else {
            setMessage('The verification link is invalid or has expired. Please try again.');
          }
        })
        .catch((error) => {
          console.error('Error verifying account:', error);
          setMessage('An error occurred during verification. Please try again.');
        })
        .finally(() => {
          setLoading(false);
        });
    } else {
      setMessage('Invalid verification link.');
      setLoading(false);
    }
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
