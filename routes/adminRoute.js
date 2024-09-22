import express from 'express';
import { getPendingComplaints, changeComplaintStatus } from '../controllers/adminController.js';
// import { adminAuth } from '../middlware/adminAuth.js';

const router = express.Router();

// router.post('/verify-login', adminAuth, verifyLogin);
router.get('/get-complaints', getPendingComplaints);
router.post('/change-status', changeComplaintStatus);

export default router;