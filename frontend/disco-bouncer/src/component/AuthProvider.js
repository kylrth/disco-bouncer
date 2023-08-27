import { createContext, useContext, useState } from 'react';

const AuthContext = createContext();

export const AuthProvider = ({ children }) => {
  const [cookies, setCookies] = useState(null);
  const [authenticated, setAuthenticated] = useState(false);

  const login = (cookies) => {
    setCookies(cookies);
    setAuthenticated(true);
  };

  const logout = () => {
    setCookies(null);
    setAuthenticated(false);
  };

  return (
    <AuthContext.Provider value={{ authenticated, cookies, login, logout }}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => useContext(AuthContext);
