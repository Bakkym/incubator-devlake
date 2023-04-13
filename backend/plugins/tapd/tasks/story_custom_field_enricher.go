/*
Licensed to the Apache Software Foundation (ASF) under one or more
contributor license agreements.  See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to You under the Apache License, Version 2.0
(the "License"); you may not use this file except in compliance with
the License.  You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package tasks

import (
	"fmt"
	"github.com/apache/incubator-devlake/core/dal"
	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/core/plugin"
	helper "github.com/apache/incubator-devlake/helpers/pluginhelper/api"
	"github.com/apache/incubator-devlake/plugins/tapd/models"
	"reflect"
)

var _ plugin.SubTaskEntryPoint = EnrichStoryCustomFields

var EnrichStoryCustomFieldMeta = plugin.SubTaskMeta{
	Name:       "enrichStoryCustomFields",
	EntryPoint: EnrichStoryCustomFields,
	// TODO false or true?
	EnabledByDefault: true,
	Description:      "Enrich story custom fields",
	DomainTypes:      []string{plugin.DOMAIN_TYPE_TICKET},
}

func EnrichStoryCustomFields(taskCtx plugin.SubTaskContext) errors.Error {
	rawDataSubTaskArgs, data := CreateRawDataSubTaskArgs(taskCtx, RAW_STORY_TABLE, false)
	db := taskCtx.GetDal()

	clauses := []dal.Clause{
		dal.From(&models.TapdStoryCustomFields{}),
		dal.Where("connection_id = ? and workspace_id = ?", data.Options.ConnectionId, data.Options.WorkspaceId),
		dal.Orderby("name ASC"),
	}

	cursor, err := db.Cursor(clauses...)
	if err != nil {
		return err
	}

	defer cursor.Close()

	converter, err := helper.NewDataConverter(helper.DataConverterArgs{
		RawDataSubTaskArgs: *rawDataSubTaskArgs,
		InputRowType:       reflect.TypeOf(models.TapdStoryCustomFields{}),
		Input:              cursor,
		Convert: func(inputRow interface{}) ([]interface{}, errors.Error) {
			customField := inputRow.(*models.TapdStoryCustomFields)

			storyCustomFieldValues := make([]*models.TapdStoryCustomFieldValue, 0)

			clausesForCustomFieldValue := []dal.Clause{
				dal.Select(fmt.Sprintf(`connection_id, workspace_id, id as story_id, %s as custom_value, '%s' as custom_field, '%s' as name`,
					customField.CustomField, customField.CustomField, customField.Name)),
				dal.From(&models.TapdStory{}),
				dal.Where("connection_id = ? and workspace_id = ?",
					data.Options.ConnectionId, data.Options.WorkspaceId),
			}

			err = db.All(&storyCustomFieldValues, clausesForCustomFieldValue...)
			results := make([]interface{}, 0, len(storyCustomFieldValues))
			for _, storyCustomFieldValue := range storyCustomFieldValues {
				results = append(results, storyCustomFieldValue)
			}
			return results, nil
		},
	})

	if err != nil {
		return err
	}

	return converter.Execute()
}