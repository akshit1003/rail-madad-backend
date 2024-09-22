import express from 'express';
import multer from 'multer';
import { submitPNR, getComplaints } from '../controllers/complaintController.js';

const router = express.Router();
const upload = multer({ storage: multer.memoryStorage() });

router.post('/submit-complaint', upload.single('image'), submitPNR);
router.get('/complaints/:pnr', getComplaints);

export default router;