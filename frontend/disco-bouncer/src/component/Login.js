import React, { useState } from 'react';
import './Login.css'; // Import your CSS file

function Login({ onLogin }) {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');

  const handleLogin = () => {
    // Perform login logic here
    // Once the login is successful, call the callback function
    onLogin();
  };

  return (
    <div className="login-page">
        <div className="logo-container">
            <img src="./logo.png" alt="Disco Bouncer Logo" className="logo" />
        </div>
        <br />
        <br />
        <div className='login-box'>
            <input
                className='input-field'
                type="text"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                placeholder="Username"
            />
            <input
                className='input-field'
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="Password"
            />
            <button onClick={handleLogin}>Login</button>
        </div>
    </div>
  );
}

export default Login;