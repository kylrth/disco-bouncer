import React, { useState } from 'react';
import Login from './Login';

function HomePage() {
  const [isLoggedIn, setIsLoggedIn] = useState(false);

  const handleLogin = () => {
    setIsLoggedIn(true);
  };

  return (
    <div className="home-page">
      {isLoggedIn ? (
        <p>This is a test</p>
      ) : (
        <Login onLogin={handleLogin} />
      )}
    </div>
  );
}

export default HomePage;