package db

import (
	"database/sql"
	"invido-site/src/idl"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func (ld *LiteDB) DeleteAllPostItem() error {
	log.Println("[LiteDB - DeleteAllPostItem] start")
	q := `DELETE FROM post;`
	if ld.debugSQL {
		log.Println("SQL is:", q)
	}

	stm, err := ld.connDb.Prepare(q)
	if err != nil {
		return err
	}
	res, err := stm.Exec()
	if ld.debugSQL {
		ra, err := res.RowsAffected()
		if err != nil {
			return err
		}
		log.Println("Row affected: ", ra)
	}
	return err
}

func (ld *LiteDB) DeletePostId(postId string) error {
	log.Println("[LiteDB - DeletePostId] delete post on post_id ", postId)
	q := `DELETE FROM post WHERE post_id=?;`
	if ld.debugSQL {
		log.Println("SQL is", q)
	}
	stmt, err := ld.connDb.Prepare(q)
	if err != nil {
		return err
	}
	res, err := stmt.Exec(postId)
	if err != nil {
		return err
	}
	if ld.debugSQL {
		ra, err := res.RowsAffected()
		if err != nil {
			return err
		}
		log.Println("Row affected: ", ra)
	}
	log.Println("[LiteDB - DeletePostId] ok")
	return nil
}

func (ld *LiteDB) UpdateMd5Post(tx *sql.Tx, postItem *idl.PostItem) error {
	log.Println("[LiteDB - UpdateMd5Post] update md5 on post id ", postItem.PostId)
	q := `UPDATE post SET md5=? WHERE post_id=?;`
	if ld.debugSQL {
		log.Println("Query is", q)
	}
	stm, err := tx.Prepare(q)
	if err != nil {
		return err
	}

	res, err := tx.Stmt(stm).Exec(postItem.Md5, postItem.PostId)
	if ld.debugSQL {
		ra, err := res.RowsAffected()
		if err != nil {
			return err
		}
		log.Println("Row affected: ", ra)
	}

	return err
}

func (ld *LiteDB) InsertNewPost(tx *sql.Tx, postItem *idl.PostItem) error {
	if ld.debugSQL {
		log.Println("[LiteDB - InsertNewPost] insert new Post on post id ", postItem.PostId)
	}

	q := `INSERT INTO post(title,post_id,timestamp,abstract,uri,title_img_uri,md5) VALUES(?,?,?,?,?,?,?);`
	if ld.debugSQL {
		log.Println("Query is", q)
	}

	stmt, err := tx.Prepare(q)
	if err != nil {
		return err
	}

	result, err := tx.Stmt(stmt).Exec(postItem.Title,
		postItem.PostId,
		postItem.DateTime.Local().Unix(),
		postItem.Abstract,
		postItem.Uri,
		postItem.TitleImgUri,
		postItem.Md5)
	if err != nil {
		return err
	}

	postItem.Id, _ = result.LastInsertId()
	if ld.debugSQL {
		log.Println("Post added into the db OK: ", postItem.Id)
	}
	return nil
}

func (ld *LiteDB) GetPostList() ([]idl.PostItem, error) {
	log.Println("[LiteDB - GetPostList] select all post")

	q := `SELECT id,title,post_id,timestamp,abstract,uri,title_img_uri,md5 from post ORDER BY post_id DESC;`
	if ld.debugSQL {
		log.Println("Query is", q)
	}
	rows, err := ld.connDb.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	res := []idl.PostItem{}
	for rows.Next() {
		var ts int64
		//  var md5 sql.NullString
		item := idl.PostItem{}
		if err := rows.Scan(&item.Id,
			&item.Title,
			&item.PostId,
			&ts,
			&item.Abstract,
			&item.Uri,
			&item.TitleImgUri,
			&item.Md5); err != nil {
			return nil, err
		}
		item.DateTime = time.Unix(ts, 0)
		item.DateTimeRfC822 = item.DateTime.Format(time.RFC822Z)
		res = append(res, item)
	}
	log.Printf("[LiteDB - GetPostList] posts read %d", len(res))
	return res, nil
}
