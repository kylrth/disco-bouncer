import React, { createContext, useContext, useState } from 'react';

const AuthContext = createContext();

export const AuthProvider = ({ children }) => {
  const [cookies, setCookies] = useState(null);
  const [authenticated, setAuthenticated] = useState(false);

  const login = (cookies) => {
    setCookies(cookies);
    setAuthenticated(true);
    localStorage.setItem('authenticated', JSON.stringify(true));
    localStorage.setItem('cookies', cookies);
  };
  
  const logout = () => {
    setCookies(null);
    setAuthenticated(false);
    localStorage.removeItem('authenticated');
    localStorage.removeItem('cookies');
  };

  return (
    <AuthContext.Provider value={{ authenticated, cookies, login, logout, setAuthenticated, setCookies }}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => useContext(AuthContext);
