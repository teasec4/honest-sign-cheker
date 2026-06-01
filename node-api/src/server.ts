import express, {
  type ErrorRequestHandler,
  type NextFunction,
  type Request,
  type Response
} from "express";
import { mkdtemp, rm, writeFile } from "node:fs/promises";
import { createServer } from "node:http";
import os from "node:os";
import path from "node:path";
import multer from "multer";
import { runDuplicates, runPrimary } from "./checker.js";

const maxUploadSize = 512 << 20;

class HttpError extends Error {
  constructor(
    readonly statusCode: number,
    message: string
  ) {
    super(message);
  }
}

const app = express();
const upload = multer({
  storage: multer.memoryStorage(),
  limits: {
    fileSize: maxUploadSize
  }
});

app.use((req, res, next) => {
  res.header("Access-Control-Allow-Origin", "*");
  res.header("Access-Control-Allow-Methods", "GET, POST, OPTIONS");
  res.header("Access-Control-Allow-Headers", "Content-Type");

  if (req.method === "OPTIONS") {
    res.sendStatus(204);
    return;
  }
  next();
});

app.get("/api/health", (_req, res) => {
  res.json({ status: "ok" });
});

app.post(
  "/api/primary-check",
  upload.fields([
    { name: "issued", maxCount: 1 },
    { name: "returned", maxCount: 1 }
  ]),
  asyncHandler(async (req, res) => {
    const minPercent = parseMinPercent(req.body?.minPercent);
    const tempDir = await mkdtemp(path.join(os.tmpdir(), "onestsignt-api-"));
    try {
      const issuedPath = await saveUploadedFile(tempDir, "issued", getUploadedFile(req, "issued"));
      const returnedPath = await saveUploadedFile(
        tempDir,
        "returned",
        getUploadedFile(req, "returned")
      );
      const report = await asBadRequest(runPrimary(issuedPath, returnedPath, minPercent));
      res.json(report);
    } finally {
      await rm(tempDir, { recursive: true, force: true });
    }
  })
);

app.post(
  "/api/duplicate-check",
  upload.single("restored"),
  asyncHandler(async (req, res) => {
    const tempDir = await mkdtemp(path.join(os.tmpdir(), "onestsignt-api-"));
    try {
      const restoredPath = await saveUploadedFile(
        tempDir,
        "restored",
        getUploadedFile(req, "restored")
      );
      const report = await asBadRequest(runDuplicates(restoredPath));
      res.json(report);
    } finally {
      await rm(tempDir, { recursive: true, force: true });
    }
  })
);

const errorHandler: ErrorRequestHandler = (error, _req, res, _next) => {
  const status =
    error instanceof HttpError ? error.statusCode : error instanceof multer.MulterError ? 400 : 500;
  const message = error instanceof Error ? error.message : "ошибка сервера";
  res.status(status).json({ error: message });
};

app.use(errorHandler);

const host = process.env.HOST ?? "127.0.0.1";
const port = Number(process.env.PORT ?? "8080");
const server = createServer(app);

server.on("error", (error) => {
  console.error("Ошибка:", error.message);
  process.exit(1);
});

server.listen(port, host, () => {
  console.log(`API-сервер запущен: http://${host}:${port}`);
});

function asyncHandler(
  handler: (req: Request, res: Response, next: NextFunction) => Promise<void>
) {
  return (req: Request, res: Response, next: NextFunction): void => {
    handler(req, res, next).catch(next);
  };
}

function parseMinPercent(raw: unknown): number {
  if (raw == null || raw === "") {
    return 85;
  }
  const value = Number(raw);
  if (!Number.isFinite(value)) {
    throw new HttpError(400, "некорректный процент совпадения");
  }
  return value;
}

function getUploadedFile(req: Request, fieldName: string): Express.Multer.File {
  if (fieldName === "restored" && req.file) {
    return req.file;
  }

  const files = req.files as Record<string, Express.Multer.File[]> | undefined;
  const file = files?.[fieldName]?.[0];
  if (!file) {
    throw new HttpError(400, `не выбран файл "${fieldName}"`);
  }
  return file;
}

async function saveUploadedFile(
  tempDir: string,
  fieldName: string,
  file: Express.Multer.File
): Promise<string> {
  const extension = path.extname(file.originalname).toLowerCase();
  if (extension === "") {
    throw new HttpError(400, `у файла "${file.originalname}" нет расширения`);
  }

  const filePath = path.join(tempDir, `${fieldName}${extension}`);
  await writeFile(filePath, file.buffer);
  return filePath;
}

async function asBadRequest<T>(promise: Promise<T>): Promise<T> {
  try {
    return await promise;
  } catch (error) {
    throw new HttpError(400, error instanceof Error ? error.message : String(error));
  }
}
