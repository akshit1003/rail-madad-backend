import admin from '../config/firebase.js';
import dotenv from 'dotenv';

dotenv.config();

const db = admin.firestore();

export const getPendingComplaints = async (req, res) => {
    try {
        const complaintsRef = admin.firestore().collection('complaints');
        const snapshot = await complaintsRef.get();

        if (snapshot.empty) {
            return res.json({ message: "No complaints found" });
        }

        let pendingComplaints = [];

        snapshot.forEach(doc => {
            const pnr = doc.id;
            const data = doc.data();
            if (data.complaints && data.complaints.length > 0) {
                const pending = data.complaints
                    .filter(complaint => complaint.status === "Pending")
                    .map(complaint => ({ ...complaint, pnr }));
                pendingComplaints = pendingComplaints.concat(pending);
            }
        });

        if (pendingComplaints.length === 0) {
            return res.json({ message: "No pending complaints found" });
        }

        res.json({
            pendingComplaints,
        });
    } catch (error) {
        console.error("Error fetching pending complaints:", error);
        res.status(500).json({ error: "Internal server error" });
    }
};

export const changeComplaintStatus = async (req, res) => {
    const { pnr, complaintId, newStatus } = req.body;

    if (!pnr || !complaintId || !newStatus) {
        return res.status(400).json({ message: 'Missing required fields' });
    }

    try {
        const complaintRef = db.collection('complaints').doc(pnr);
        const doc = await complaintRef.get();

        if (!doc.exists) {
            return res.status(404).json({ message: 'Complaint not found' });
        }

        const data = doc.data();
        const updatedComplaints = data.complaints.map(complaint => {
            if (complaint.id === complaintId) {
                return { ...complaint, status: newStatus };
            }
            return complaint;
        });

        await complaintRef.update({ complaints: updatedComplaints });

        res.status(200).json({
            message: 'Complaint status updated successfully',
            pnr,
            complaintId,
            newStatus
        });
    } catch (error) {
        console.error('Error updating complaint status:', error);
        res.status(500).json({ message: 'Internal server error', error: error.message });
    }
};