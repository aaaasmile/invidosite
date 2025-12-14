package db

import (
	"database/sql"
	"fmt"
	"invido-site/src/idl"
	"log"
	"strings"
	"time"
)

func (ld *LiteDB) GetTagPostMap(tags []idl.TagItem) (map[string][]*idl.PostItem, error) {
	log.Println("[LiteDB - GetTagPostMap] map all tags")
	res := map[string][]*idl.PostItem{}
	for _, tag_key := range tags {
		lst, err := ld.getPostlistWithTag(tag_key.Title)
		if err != nil {
			return nil, err
		}
		res[tag_key.Title] = lst
	}
	return res, nil
}

func (ld *LiteDB) getPostlistWithTag(tag_title string) ([]*idl.PostItem, error) {
	res := []*idl.PostItem{}
	q := `SELECT p.*
			FROM post p
			INNER JOIN tags_to_post ttp ON p.id = ttp.post_id
			INNER JOIN tags t ON ttp.tag_id = t.id
			WHERE t.title = ?;`
	if ld.debugSQL {
		log.Println("Query is", q)
	}
	stm, err := ld.connDb.Prepare(q)
	if err != nil {
		return nil, err
	}
	rows, err := stm.Query(tag_title)
	if err != nil {
		return nil, err
	}
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
		res = append(res, &item)
	}
	log.Printf("[LiteDB - getPostlistWithTag] posts read %d with tag %s", len(res), tag_title)
	return res, nil
}

func (ld *LiteDB) GetTagList() ([]idl.TagItem, error) {
	log.Println("[LiteDB - GetTagList] select all tags")

	q := `SELECT id,title,timestamp,uri,md5,numofposts from tags ORDER BY title DESC;`
	if ld.debugSQL {
		log.Println("Query is", q)
	}
	rows, err := ld.connDb.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	res := []idl.TagItem{}
	for rows.Next() {
		var ts int64
		item := idl.TagItem{}
		if err := rows.Scan(&item.Id,
			&item.Title,
			&ts,
			&item.Uri,
			&item.Md5,
			&item.NumOfPosts); err != nil {
			return nil, err
		}
		item.DateTime = time.Unix(ts, 0)
		item.DateTimeRfC822 = item.DateTime.Format(time.RFC822Z)
		res = append(res, item)
	}
	log.Printf("[LiteDB - GetTagList] tags read %d", len(res))
	return res, nil
}

func (ld *LiteDB) InsertOrUpdateTag(tx *sql.Tx, tag string, postItem *idl.PostItem) (bool, error) {
	tag = strings.Trim(tag, " ")
	if ld.debugSQL {
		log.Println("[LiteDB - InsertOrUpdateTag] insert or update tag on post id ", tag, postItem.PostId)
	}

	q := `SELECT EXISTS( 
	         SELECT 1 FROM tags
			 WHERE title = ?
		  );`
	if ld.debugSQL {
		log.Println("Query is", q)
	}

	stmt, err := tx.Prepare(q)
	if err != nil {
		return false, err
	}
	var exists bool
	err = stmt.QueryRow(tag).Scan(&exists)
	if err != nil {
		return false, err
	}
	tagItem := idl.TagItem{Title: tag}
	if !exists {
		if err := ld.insertTagInTags(tx, &tagItem); err != nil {
			return false, err
		}
	} else {
		if err := ld.selectTag(tx, &tagItem); err != nil {
			return false, err
		}
	}

	return ld.insertOrUpdateTagsToPostId(tx, &tagItem, postItem)
}

func (ld *LiteDB) selectTag(tx *sql.Tx, tagItem *idl.TagItem) error {
	q := `SELECT id FROM tags WHERE title = ?`
	if ld.debugSQL {
		log.Println("Query is", q)
	}
	stmt, err := tx.Prepare(q)
	if err != nil {
		return err
	}

	var id int64
	err = stmt.QueryRow(tagItem.Title).Scan(&id)
	if err != nil {
		return err
	}
	tagItem.Id = id
	return nil
}

func (ld *LiteDB) insertOrUpdateTagsToPostId(tx *sql.Tx, tagItem *idl.TagItem, postItem *idl.PostItem) (bool, error) {
	q2 := `SELECT EXISTS( 
	         SELECT 1 FROM tags_to_post
			 WHERE tag_title = ? AND post_id_txt = ? 
		  );`
	if ld.debugSQL {
		log.Println("Query is", q2)
	}

	stmt2, err := tx.Prepare(q2)
	if err != nil {
		return false, err
	}
	var exists2 bool
	err = stmt2.QueryRow(tagItem.Title, postItem.PostId).Scan(&exists2)
	if err != nil {
		return false, err
	}
	if !exists2 {
		if err := ld.insertTagInTagsToPost(tx, tagItem, postItem); err != nil {
			return false, err
		}
		return true, nil
	} else {
		return false, nil
	}
}

func (ld *LiteDB) insertTagInTagsToPost(tx *sql.Tx, tagItem *idl.TagItem, postItem *idl.PostItem) error {
	if ld.debugSQL {
		log.Println("[LiteDB - insertTagInTagsToPost] insert new Tag in Tags_to_post ", tagItem.Title, postItem.PostId)
	}

	q := `INSERT INTO tags_to_post(post_id_txt,tag_title,tag_id,post_id) VALUES(?,?,?,?);`
	if ld.debugSQL {
		log.Println("Query is", q)
	}

	stmt, err := tx.Prepare(q)
	if err != nil {
		return err
	}
	if tagItem.Id == 0 {
		return fmt.Errorf("[insertTagInTagsToPost] tagItem.Id  is not set")
	}
	if postItem.Id == 0 {
		return fmt.Errorf("[insertTagInTagsToPost] postItem.Id is not set")
	}

	result, err := tx.Stmt(stmt).Exec(postItem.PostId,
		tagItem.Title,
		tagItem.Id,
		postItem.Id)
	if err != nil {
		return err
	}

	if ld.debugSQL {
		id, _ := result.LastInsertId()
		log.Println("Tag added into the db OK: ", id)
	}
	return nil
}

func (ld *LiteDB) insertTagInTags(tx *sql.Tx, tagItem *idl.TagItem) error {
	if ld.debugSQL {
		log.Println("[LiteDB - insertTagInTags] insert new Tag in Tags ", tagItem.Title)
	}
	if tagItem.Title == "" {
		return fmt.Errorf("[insertTagInTags] Tag is empty")
	}

	q := `INSERT INTO tags(title,timestamp,uri,md5,numofposts) VALUES(?,?,?,?,?);`
	if ld.debugSQL {
		log.Println("Query is", q)
	}

	stmt, err := tx.Prepare(q)
	if err != nil {
		return err
	}
	uri := fmt.Sprintf("/tags/%s/#", tagItem.Title)
	timeNow := time.Now()
	md5 := " "
	num_of_posts := 0
	result, err := tx.Stmt(stmt).Exec(tagItem.Title,
		timeNow.Local().Unix(),
		uri,
		md5,
		num_of_posts)
	if err != nil {
		return err
	}

	tagItem.Id, _ = result.LastInsertId()
	if ld.debugSQL {
		log.Println("Tag added into the db OK: ", tagItem.Id)
	}
	return nil
}

func (ld *LiteDB) DeleteAllTags() error {
	log.Println("[LiteDB - DeleteAllTags] start")
	q := `DELETE FROM tags;`
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

func (ld *LiteDB) DeleteTagOnPostId(postId string) error {
	log.Println("[LiteDB - DeleteAllTagsToPost] start")
	q := `DELETE FROM tags_to_post WHERE post_id=?;`
	if ld.debugSQL {
		log.Println("SQL is:", q)
	}

	stm, err := ld.connDb.Prepare(q)
	if err != nil {
		return err
	}
	res, err := stm.Exec(postId)
	if ld.debugSQL {
		ra, err := res.RowsAffected()
		if err != nil {
			return err
		}
		log.Println("Row affected: ", ra)
	}
	return err
}

func (ld *LiteDB) DeleteAllTagsToPost() error {
	log.Println("[LiteDB - DeleteAllTagsToPost] start")
	q := `DELETE FROM tags_to_post;`
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

func (ld *LiteDB) UpdateNumOfPostInTags() error {
	lst, err := ld.GetTagList()
	if err != nil {
		return err
	}
	for _, tag := range lst {
		tag.NumOfPosts, err = ld.getNumOfPostInTag(&tag)
		if err != nil {
			return err
		}
		if err := ld.updateNumOfPostsInTag(&tag); err != nil {
			return err
		}
	}
	return nil
}

func (ld *LiteDB) getNumOfPostInTag(tagItem *idl.TagItem) (int, error) {
	log.Println("[LiteDB - getNumOfPostInTag] get num of posts ", tagItem.Title)
	q := `SELECT COUNT(*) FROM tags_to_post WHERE tag_title = ?;`
	if ld.debugSQL {
		log.Println("Query is", q)
	}
	stm, err := ld.connDb.Prepare(q)
	if err != nil {
		return 0, err
	}
	var count int
	err = stm.QueryRow(tagItem.Title).Scan(&count)

	return count, err
}

func (ld *LiteDB) updateNumOfPostsInTag(tagItem *idl.TagItem) error {
	log.Println("[LiteDB - updateNumOfPostsInTag] update num of posts ", tagItem.NumOfPosts)
	if tagItem.NumOfPosts == 0 {
		return fmt.Errorf("[updateNumOfPostsInTag] num of posts is zero")
	}
	q := `UPDATE tags SET numofposts=? WHERE id=?;`
	if ld.debugSQL {
		log.Println("Query is", q)
	}
	stm, err := ld.connDb.Prepare(q)
	if err != nil {
		return err
	}

	res, err := stm.Exec(tagItem.NumOfPosts, tagItem.Id)
	if ld.debugSQL {
		ra, err := res.RowsAffected()
		if err != nil {
			return err
		}
		log.Println("Row affected: ", ra)
	}

	return err
}
