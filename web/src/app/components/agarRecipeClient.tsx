'use client'

import React, {JSX, useState} from "react";
import {IsValidNote, NewEntryNotes, Note, NotesFormArea} from "@/app/components/formSubcomponents/notes";
import {
    AddCreatedTriColFunction,
    AllEntries,
    OnViewCreatorQuadCol
} from "@/app/components/formSubcomponents/shared";
import ID from "@/app/components/formSubcomponents/id";
import DateArea from "@/app/components/formSubcomponents/date";
import {AgarRecipeData} from "@/app/components/agarRecipeServer";
import LiquidsArea, {
    IsValidLiquid,
    Liquid,
    LiquidEntriesGroupForNew
} from "@/app/components/formSubcomponents/liquids";
import NutrientsArea, {
    IsValidNutrient,
    Nutrient,
    NutrientsEntriesGroupForNew,
} from "@/app/components/formSubcomponents/nutrients";
import SugarsArea, {
    IsValidSugar,
    Sugar,
    SugarEntriesGroupForNew,
} from "@/app/components/formSubcomponents/sugars";
import {
    createApiUrlFor,
    CreatedLinkFor,
    CreateNewEntryButton, dataFor, DisplayFormWrapper,
    DisplayInput, DoCreateRequest, DoUpdateRequest, ErrHandler, ExistingDualSelector, FlexedArea, FlexedSinglesGroup,
    HandleJsonResponse,
    IsString,
    ListPageItems, ListPageTable, ListTableColumn, NewColumn, NewEntryFormWrapper,
    NewEntryInput, NumberToDateStr,
    OptionalArrayOfType,
    OptionalKey,
    RequiredArrayOfType, updateApiUrlFor,
    ViewInNewTabButton
} from "@/app/components/common";
import EntryLink, {EntryLinkWrapper} from "@/app/components/formSubcomponents/entryLink";
import {ErrorDisplay, InlineTitle, NameArea, StandardArea} from "@/app/components/formSubcomponents/commonClient";
import {BaseExternalUrl} from "@/app/components/Constants";
import AdditivesArea, {
    Additive,
    AdditiveEntriesGroupForNew,
    IsValidAdditive
} from "@/app/components/formSubcomponents/additives";
import {
    Antibiotic,
    AntibioticEntriesGroupForNew,
    AntibioticsDisplay,
} from "@/app/components/formSubcomponents/antibiotic";
import TestAndValidate from "@/app/components/testing/untested";
import {AclDisplay, IsValidAcl, MarshalAcl, TogglableAreaWithDepth} from "@/app/components/accessControlClient";
import {ACL} from "@/app/components/accessControlServer";
import {AssertAgarBatch, NewAgarBatchForm} from "@/app/components/agarBatchClient";
import {AgarBatchData} from "@/app/components/agarBatchServer";
import {InputNumber} from "@/app/components/formSubcomponents/numericInput";
import {OnViewCreatorsTriColArea} from "@/app/components/formSubcomponents/ovc";
import {InitialNotesState} from "@/app/components/formSubcomponents/initialState";

export function AssertAgarRecipe(input: any): asserts input is AgarRecipeData {
    if (typeof input !== 'object') {
        throw new Error('Input is not an object! Input is ' + typeof input);
    }
    // required simple keys
    let requiredSimpleKeys = new Map<string, string>([
        ['_id', 'string'],
        ['name', 'string'],
        ['agar', 'number'],
        ['standard', 'boolean'],
        ['lastUpdated', 'number'],
    ])
    for (let [key, expType] of requiredSimpleKeys) {
        if (!(key in input && typeof input[key] === expType)) {
            console.error('Agar Recipe assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
            console.error(JSON.stringify(input));
            console.error(JSON.stringify(input[key]));
            throw new Error('Agar Recipe assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
        }
    }

    // complex optional simple keys
    let complexOptionalKeys = new Map<string, (v: any) => boolean>([
        ['acl', IsValidAcl]
    ])
    for (let [key, validator] of complexOptionalKeys) {
        if (!OptionalKey(key, input, validator)) {
            console.error('AgarRecipe assertion failure: optional key ' + key + ' was not valid');
            console.error(JSON.stringify(input));
            console.error(JSON.stringify(input[key]));
            throw new Error('AgarRecipe assertion failure: optional key ' + key + ' was not valid');
        }
    }

    // complex required array keys
    let complexRequiredArrayKeys = new Map<string, (v: any) => boolean>([
        ['liquids', IsValidLiquid],
    ])
    for (let [key, validator] of complexRequiredArrayKeys) {
        if (!RequiredArrayOfType(key, input, validator)) {
            console.error('AgarRecipe assertion failure: required array key ' + key + ' was not valid');
            console.error(JSON.stringify(input));
            console.error(JSON.stringify(input[key]));
            throw new Error('AgarRecipe assertion failure: required array key ' + key + ' was not valid');
        }
    }

    // complex optional array keys
    let complexOptionalArrayKeys = new Map<string, (v: any) => boolean>([
        ['nutrients', IsValidNutrient],
        ['sugars', IsValidSugar],
        ['additives', IsValidAdditive],
        ['antibiotics', IsString],
        ['notes', IsValidNote],
    ])
    for (let [key, validator] of complexOptionalArrayKeys) {
        if (!OptionalArrayOfType(key, input, validator)) {
            console.error('AgarRecipe assertion failure: optional array key ' + key + ' was not valid');
            console.error(JSON.stringify(input));
            console.error(JSON.stringify(input[key]));
            throw new Error('AgarRecipe assertion failure: optional array key ' + key + ' was not valid');
        }
    }
    return
}

export default function AgarRecipeDisplay(
    {
        id, readonly, data, headerLevel, isTopLevel, cookies
    }:
    DisplayInput
) {
    try {
        AssertAgarRecipe(data)
        const [initial, setInitial] = useState(data)
        // Required
        const [name, setName] = useState(data.name)
        // Optional
        const [isStandard, setIsStandard] = useState(data.standard)
        const [notes, setNotes] = useState<AllEntries<Note>>({existing: dataFor(data.notes || []), new: []})
        const [acl, setAcl] = useState<ACL | undefined>(data.acl)
        const [err, setErr] = useState<string | undefined>()
        const updateInitial = (updated: AgarRecipeData) => {
            setInitial(updated)
            setName(updated.name)
            setIsStandard(updated.standard)
            setNotes(InitialNotesState(updated.notes))
            setAcl(updated.acl)
        }

        const agarRecipeSubmit = () => {
            if (name === undefined || name === "") {
                setErr("Name field must not be empty")
                return
            }
            const body: any = {
                name: name,
                standard: isStandard,
                notes: notes,
                acl: MarshalAcl(acl), // TODO; use this everywhere if it works
            }
            DoUpdateRequest("agarRecipe",initial._id, body, AssertAgarRecipe)
                .then(updateInitial)
                .catch(ErrHandler(setErr))
            // fetch(updateApiUrlFor("agarRecipe",initial._id), {
            //     method: 'Post',
            //     body: JSON.stringify({
            //         name: name,
            //         standard: isStandard,
            //         notes: notes,
            //         acl: MarshalAcl(acl), // TODO; use this everywhere if it works
            //     }),
            //     headers: clientPostRequestHeaders,
            // })
            //     .then(HandleJsonResponse)
            //     .then((newEntry) => {
            //         try {
            //             AssertAgarRecipe(newEntry)
            //             updateInitial(newEntry)
            //         } catch (er) {
            //             throw new Error("failed to decode response:" + JSON.stringify(er))
            //         }
            //     })
                //.catch(ErrHandler(setErr));
        }
        const ovcs: OnViewCreatorQuadCol[] = [
            {
                txt: "Create Batch From Recipe",
                newCreationArea: (onCreate: AddCreatedTriColFunction) => {
                    return <NewAgarBatchForm agarRecipeIn={data} handlers={{
                        onCreate: (newItem: AgarBatchData) => {
                            return onCreate([{
                                typeText: "Agar Batch",
                                node: <CreatedLinkFor linkId={newItem._id} typ={"agarBatch"}/>
                            }], false)
                        },
                        isTopLevel: false,
                    }}/>
                },
            }
        ]
        return (
            <DisplayFormWrapper entryType={"agarRecipe"}>
                <ErrorDisplay err={err} headerLevel={headerLevel}/>
                <TestAndValidate todos={["Put name at top????"]}>
                    <ID id={data._id} txt={"Agar Recipe"} entryType={"agarRecipe"}/>
                </TestAndValidate>
                <OnViewCreatorsTriColArea OnViewCreators={ovcs} readonly={readonly}/>{/* TODO: where to put?*/}
                <FlexedArea>
                    <FlexedSinglesGroup>
                        <NameArea currentName={name} setName={setName}
                                  readonly={readonly}/>{/*TODO: Allow changing??? Make this area longer!*/}
                    </FlexedSinglesGroup>
                    <FlexedSinglesGroup>
                        <StandardArea isStandard={isStandard} setStandard={setIsStandard} readonly={readonly}
                                      headerLevel={headerLevel}/>{/* TODO: upon change (only when clicked first), deletes users from ACL. FIX THAT*/}
                        <div className={"inlineChildren"}>
                            <div>{"Agar g/L: "}</div>
                            <div>{initial.agar}</div>
                        </div>
                        <DateArea pre={"Last Updated: "} when={initial.lastUpdated} readonly={true}/>
                    </FlexedSinglesGroup>
                </FlexedArea>

                <LiquidsArea initialValues={dataFor(initial.liquids)}
                             readonly={true}/>{/* TODO: FIX AND REFORMAT THIS*/}
                <NutrientsArea initialValues={dataFor(initial.nutrients)}
                               readonly={true}/>{/* TODO: FIX AND REFORMAT THIS*/}
                <SugarsArea initialValues={dataFor(initial.sugars)} readonly={true}/>{/* TODO: FIX AND REFORMAT THIS*/}
                <AdditivesArea readonly={true}
                               initialValues={dataFor(initial.additives)}/>{/* TODO: FIX AND REFORMAT THIS*/}
                <AntibioticsDisplay antibiotics={initial.antibiotics}/>{/* TODO: FIX AND REFORMAT THIS*/}
                <NotesFormArea readonly={readonly} initial={initial.notes} updateParent={setNotes}/>
                <TogglableAreaWithDepth startOpen={false} openTxt={"view permissions"} closeTxt={"minimize perms area"}>
                    <AclDisplay ACL={acl} readonly={readonly} updateParent={setAcl}/>
                </TogglableAreaWithDepth>
                {/* TODO: fix the ADD A PROJECT area*/}
                {readonly ? null :
                    <button className={"bottomButton greenButton"} onClick={(e) => {
                        e.stopPropagation();
                        agarRecipeSubmit()
                    }}>{"Update"}</button>}

            </DisplayFormWrapper>
        )
    } catch (err) {
        return <div>{"ERROR: Agar Recipe data format incorrect: " + err}</div>
    }
}

export function NewAgarRecipeForm({handlers}: { handlers: NewEntryInput<AgarRecipeData> }) {
    const [name, setName] = useState("")
    const [isStandard, setIsStandard] = useState(false)
    const [agar, setAgar] = useState(20)
    const [liquids, setLiquids] = useState<Liquid[]>([])
    const [nutrients, setNutrients] = useState<Nutrient[]>([])
    const [sugars, setSugars] = useState<Sugar[]>([])
    const [additives, setAdditives] = useState<Additive[]>([])
    const [antibiotics, setAntibiotics] = useState<Antibiotic[]>([])
    const [notes, setNotes] = useState<Note[]>([])
    const [err, setErr] = useState<string | undefined>()
    const [agarErr, setAgarErr] = useState<string | undefined>()
    const [templateSelectorOpen, setTemplateSelectorOpen] = useState<boolean>(false)
    const errHandler = ErrHandler(setErr)
    const newAgarRecipeSubmit = () => {
        if (name === "") {
            setErr("name must not be empty")
            return
        }
        if (liquids.length === 0) {
            setErr("at least one liquid must exist")
            return
        }
        let body: any = {
            name: name,
            standard: isStandard,
            agar: agar,
            liquids: liquids,
            // Optional
            nutrients: nutrients.length !== 0 ? nutrients : undefined,
            sugars: sugars.length !== 0 ? sugars : undefined,
            additives: additives.length !== 0 ? additives : undefined,
            antibiotics: antibiotics.length !== 0 ? antibiotics : undefined,
            notes: notes.length !== 0 ? notes : undefined,
        }
        DoCreateRequest("agarRecipe", body, AssertAgarRecipe)
            .then(handlers?.onCreate)
            .catch(errHandler)
        // fetch(createApiUrlFor("agarRecipe"), {
        //     method: 'Post',
        //     body: JSON.stringify(body),
        //     headers: clientPostRequestHeaders,
        // }).then(HandleJsonResponse).then((newRecipe) => {
        //     try {
        //         AssertAgarRecipe(newRecipe)
        //         handlers.onCreate && handlers.onCreate(newRecipe)
        //     } catch (e) {
        //         setErr("result was not recipe: " + JSON.stringify(e))
        //     }
        // })
            //.catch(ErrHandler(setErr));
    }
    const templateRecipeSelector = () => {
        if (templateSelectorOpen) {
            return <AgarRecipeSelector doSelect={(rec) => {
                if (rec === undefined) {
                    return
                }
                setName(rec.name)
                setIsStandard(rec.standard)
                setAgar(rec.agar)
                setLiquids(rec.liquids)
                setNutrients(rec.nutrients || [])
                setSugars(rec.sugars || [])
                setAdditives(rec.additives || [])
                setAntibiotics(rec.antibiotics || [])
                // notes?
                setTemplateSelectorOpen(false)
            }}/>
        } else {
            return <button className={"basicButton"} onClick={() => {
                setTemplateSelectorOpen(true)
            }}>{"Select a template recipe (optional)"}</button>
        }
    }
    return (
        <NewEntryFormWrapper entryType={"agarRecipe"}>
            <ErrorDisplay err={err}/>
            <div>
                {templateRecipeSelector()}
            </div>
            <NameArea classNames={"inlineChildren"} titleClasses={"mr-2"} currentName={name || ""} setName={setName}
                      headerTxt={"Recipe Name: "} readonly={false}/>
            <StandardArea isStandard={isStandard} setStandard={setIsStandard} readonly={false}
                          headerTxt={"Standard Recipe? "}/>
            <div className={"inlineChildren my-4"}>
                <InlineTitle title={"Agar g/L: "} titleClasses={"mr-2"}>
                    <InputNumber readonly={false} value={"" + agar} min={0} max={100}
                                 mode={"integer"} onChange={(e) => {
                        const val = e || ""
                        try {
                            if (val === '' || /^\d*\.?\d*$/.test(val)) {
                                setAgar(parseFloat(val));
                            }
                        } catch {
                            setAgarErr("invalid agar amount")
                        }
                    }} placeholder={"10"} errorMessage={agarErr}/>
                    <div className={"ml-2"}>
                        {agarPer400mL(agar)}
                    </div>
                </InlineTitle>
            </div>
            <div>
                <div>{"Liquids: "}</div>
                <LiquidEntriesGroupForNew currentEntries={liquids} updateParent={setLiquids}/>
            </div>
            <div>
                <div>{"Nutrients: "}</div>
                <NutrientsEntriesGroupForNew currentEntries={nutrients}
                                             updateParent={setNutrients}/>
            </div>
            <div>
                <div>{"Sugars: "}</div>
                <SugarEntriesGroupForNew currentEntries={sugars} updateParent={setSugars}/>
            </div>
            <div>
                <div>{"Additives: "}</div>
                <AdditiveEntriesGroupForNew currentEntries={additives} updateParent={setAdditives}/>
            </div>
            <div>
                <div>{"Antibiotics: "}</div>
                <AntibioticEntriesGroupForNew currentEntries={antibiotics}
                                              updateParent={setAntibiotics}/>
            </div>
            <NewEntryNotes setNotes={setNotes}/>
            {/* SUBMIT AREA */}
            <CreateNewEntryButton onSubmit={newAgarRecipeSubmit}/>
        </NewEntryFormWrapper>
    )
}

export function agarPer400mL(agar: number) {
    return <div>{"(" + (agar * 2.0 / 5.0) + " g/400mL)"}</div>
}

export const AgarRecipeArea = ({agarRecipeBinId}: { agarRecipeBinId?: string }) => {
    let linkArea: JSX.Element | null = <div>{"unknown"}</div>
    if (agarRecipeBinId !== undefined) {
        const displayId = agarRecipeBinId
        linkArea = <EntryLink
            props={{displayedId: displayId, linkId: displayId, entryType: "agarRecipe"}}> {/* TODO: DISPLAY NAME? */}
            <div>{displayId}</div>
            {/* TODO: NAME? */}
        </EntryLink>
    }
    return <div className={"agarRecipeArea"}>
        <div>{"Agar Recipe ID: "}</div>
        <div>{linkArea}</div>
    </div>
}


export function AgarRecipeSelector(
    {
        doSelect,
        allowCreate
    }: {
        doSelect: (val: AgarRecipeData | undefined) => void,
        allowCreate?: boolean
    }) {
    const table = (items: AgarRecipeData[]):JSX.Element=>{
        return <AgarRecipeSelectorTable data={items} onClick={doSelect}
                                          withLink={true}/>
    }

    return <ExistingDualSelector entryType={"agarRecipe"} entryTypes={"agarRecipes"} doSelect={doSelect} asserter={AssertAgarRecipe}
                                 table={table}>
        {allowCreate && <NewAgarRecipeForm handlers={{onCreate: doSelect,isTopLevel: false}}/>}
    </ExistingDualSelector>
}

export function AgarRecipeListPageTable({data, onClick, withLink}: ListPageItems<AgarRecipeData>) {
    let cols: ListTableColumn<AgarRecipeData>[] = [
        NewColumn("ID", (v) => v._id),
        NewColumn("Name", (v) => v.name), // TODO: shortname?
        NewColumn("Liquids", (v) => {
            return <div>
                {v.liquids.map((l, i) => {
                    return <div key={l.name + i}>{l.name}</div>
                })}
            </div>
        }),
        NewColumn("Nutrients", (v) => {
            return <div>
                {v.nutrients && v.nutrients.map((v, i) => {
                    return <div key={v.nutrient + i}>{v.nutrient}</div>
                })}
            </div>
        }),
        NewColumn("Sugars", (v) => {
            return <div>
                {v.sugars && v.sugars.map((v, i) => {
                    return <div key={v.type + i}>{v.type}</div>
                })}
            </div>
        }),
        NewColumn("Additives", (v) => {
            return <div>
                {v.additives && v.additives.map((v, i) => {
                    return <div key={v.additive + i}>{v.additive}</div>
                })}
            </div>
        }),
        NewColumn("Antibiotics", (v) => {
            return <div>
                {v.antibiotics && v.antibiotics.map((v, i) => {
                    return <div key={i}>{v}</div>
                })}
            </div>
        }),
        NewColumn("Last Updated", (v) => {
            return NumberToDateStr(v.lastUpdated)
        })

    ]
    if (withLink) {
        cols = [...cols, NewColumn("Link", (v: AgarRecipeData) => {
            return <EntryLinkWrapper props={{linkId: encodeURI(v._id), entryType: "agarRecipe", openInNewTab: true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })]
    }
    return <ListPageTable className={"text-xs"} cols={cols} data={data} onClick={onClick}/>
}

export function AgarRecipeSelectorTable({data, onClick, withLink}: ListPageItems<AgarRecipeData>) {
    let cols: ListTableColumn<AgarRecipeData>[] = [
        NewColumn("Name", (v) => v.name), // TODO: shortname?
        NewColumn("ID", (v) => v._id),
        NewColumn("Last Updated", (v) => {
            return NumberToDateStr(v.lastUpdated)
        })
    ]
    if (withLink) {
        cols = [...cols, NewColumn("Link", (v: AgarRecipeData) => {
            return <ViewInNewTabButton entryType={"agarRecipe"} id={v._id}/>
        })]
    }
    return <ListPageTable className={"text-xs"} cols={cols} data={data} onClick={onClick}/>
}
