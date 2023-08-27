import React from 'react';
// import './ManageUser.css';
import axios from 'axios';
import { Link, useNavigate } from 'react-router-dom';

function ManageUser({ isLoggedIn, username }) {
  const navigate = useNavigate();
  
  // const [newPassword, setNewPassword] = useState('');

  const handleLogout = async () => {
    try {
      const response = await axios.post('https://discobouncer.kylrth.com/logout'); // Make a POST request to /logout
      if (response.status === 200) {
        isLoggedIn = false;
        navigate('/login');
      } else {
        // Handle error case if needed
      }
    } catch (error) {
      // Handle error case if needed
    }
  };

  // const handleChangePassword = () => {
  //   // Perform the logic to change the password using newPassword state
  //   console.log(`Changing password for user ${username} to: ${newPassword}`);
  //   // You can make an API call here to update the user's password
  // };

  return (
    <div className="manage-user">
      <nav className="top-navbar">
        <Link to="/manage-user" className="top-navbar-link">Change Password</Link>
        <Link to="/logout" className="top-navbar-link" onClick={handleLogout}>Logout</Link>
      </nav>
      <h2>Manage User: {username}</h2>
      <br />
      <p>This page is under construction</p>
      {/* <div className="password-change">
        <label>New Password:</label>
        <input
          type="password"
          value={newPassword}
          onChange={(e) => setNewPassword(e.target.value)}
        />
        <button onClick={handleChangePassword}>Change Password</button>
        <br />
        <br />
        <Link to="/home">Back</Link>
      </div> */}
    </div>
  );
}

export default ManageUser;
