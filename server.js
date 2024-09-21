import express from 'express'
import cors from 'cors'
import dotenv from 'dotenv'

dotenv.config();

const app = express();
const port = process.env.PORT || 5000;

const allowedOrigins = ['http://localhost:3000'];

const corsOptions = {
    origin: function (origin, callback) {
        if (allowedOrigins.indexOf(origin) !== -1 || !origin) {
            callback(null, true);
        } else {
            callback(new Error('Not allowed by CORS'));
        }
    },
    credentials: true,
};

app.use(cors(corsOptions));

app.get('/', (req, res) => {
    res.send('Hello World!');
});

app.listen(port, () => {
    console.log(`http://localhost:${port}`);
});
