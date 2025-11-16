import { Database } from "bun:sqlite";



export class ContactFormTable {
    db: Database;
    tableName: string = "contact_message";
    constructor(db: Database) {
        this.db = db;
        this.initTable()
    }
    initTable(): void {
        const sql = `
            CREATE TABLE IF NOT EXISTS ${this.tableName} (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                name TEXT NOT NULL,
                email TEXT NOT NULL,
                message TEXT NOT NULL,
                created_at DATETIME DEFAULT CURRENT_TIMESTAMP
            );
        `;
        this.db.run(sql);
    }
    insert(name: string, email: string, message: string) {
      const stmt = this.db.prepare(
          `INSERT INTO ${this.tableName} (name, email, message) VALUES (?, ?, ?)`
      );
      const result = stmt.run(name, email, message);
      return { lastInsertRowid: Number(result.lastInsertRowid), changes: result.changes };
    }

}