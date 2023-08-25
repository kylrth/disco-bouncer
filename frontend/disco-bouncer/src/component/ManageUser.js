import React, { useState } from 'react';
import './ManageUser.css';

function ManageUser({ username }) {
  const [newPassword, setNewPassword] = useState('');

  const handleChangePassword = () => {
    // Perform the logic to change the password using newPassword state
    console.log(`Changing password for user ${username} to: ${newPassword}`);
    // You can make an API call here to update the user's password
  };

  return (
    <div className="manage-user">
      <h2>Manage User: {username}</h2>
      <div className="password-change">
        <label>New Password:</label>
        <input
          type="password"
          value={newPassword}
          onChange={(e) => setNewPassword(e.target.value)}
        />
        <button onClick={handleChangePassword}>Change Password</button>
      </div>
    </div>
  );
}

export default ManageUser;
