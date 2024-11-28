import React, { useEffect, useState } from 'react';
import PropTypes from 'prop-types';
import Cookies from 'js-cookie';
import { useNavigate } from 'react-router-dom';
import Context from './LoggedInContext';

const LoggedInProvider = ({ children }) => {
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [isLoginDone, setIsLoginDone] = useState(false);
  const [accessToken, setAccessToken] = useState('');
  const [sessionID, setSessionID] = useState('');
  const navigate = useNavigate();

  useEffect(() => {
    const acxsToken = Cookies.get('access_token');
    const session = Cookies.get('session');
    console.log("access token and session ", acxsToken, session);
    if (acxsToken && session) {
      setAccessToken(acxsToken);
      setSessionID(session);
      setIsLoggedIn(true);
    }
    setIsLoginDone(true);
  }, []);

  useEffect(() => {
    if (isLoggedIn) {
      setAccessToken(Cookies.get('access_token'));
      setSessionID(Cookies.get('session'));
    }
  }, [isLoggedIn]);

  const goToLogout = () => {
    navigate("/logout");
  };

  const clearStorageAndContextOnLogout = () => {
    localStorage.clear();
    sessionStorage.clear();
  };

  const login = () => {
    // saveToken(tokenData);
    setIsLoggedIn(true);
  };

  const logout = () => {
    setIsLoggedIn(false);
  };

  const logoutOnTokenExpiry = () => {
    goToLogout();
    clearStorageAndContextOnLogout();
  };

  return (
    <Context.Provider
      value={{ isLoggedIn, login, logout, accessToken, sessionID, logoutOnTokenExpiry, isLoginDone }}
    >
      {children}
    </Context.Provider>
  );
};

LoggedInProvider.propTypes = {
  children: PropTypes.element,
};

LoggedInProvider.defaultProps = {
  children: null,
};

export default LoggedInProvider;
