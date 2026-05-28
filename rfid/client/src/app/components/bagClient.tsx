'use client'

import React, {JSX, useState} from "react";
import {IsValidNote, NewEntryNotes, Note, NotesFormArea} from "@/app/components/formSubcomponents/notes";
import {
    AllEntries,
    OnViewCreatorQuadCol,
    SplitAllEntries
} from "@/app/components/formSubcomponents/shared";
import ID from "@/app/components/formSubcomponents/id";
import DateArea from "@/app/components/formSubcomponents/date";
import {BagData} from "@/app/components/bagServer";
import {
    InitialPicsEntries,
    IsValidPicWithNotesIncoming,
    NewPicWithNotesForm,
    PicWithNotesForm,
} from "@/app/components/formSubcomponents/picWithNotes";
import {InnocDisplay, TransfersOutDisplay} from "@/app/components/transferClient";
import {
    DisplayFormWrapper,
    DisplayInput, ExistingRecentSelector, FlexedArea, FlexedSinglesGroup,
    HandleJsonResponse,
    HandleTxtResponse,
    ImportDisplayInput, ImportEntryFormWrapper,
    ListPageItems, ListPageTable, ListTableColumn, NewColumn, NewEntryFormWrapper,
    NewEntryInput, NumberToDateStr,
    OptionalArrayOfType,
    OptionalKey,
    OptionalSimpleKey,
    RequiredKey,
    resolveContamsFormData,
    resolvePicsFormData, SelectorWrapper,
    SendMultipartRequest,
    setFormData,
    setFormImages
} from "@/app/components/common";
import {
    DisposedDisplay,
    ErrorDisplay,
    GensFormDisplay,
    MostRecentImageDisplay,
    ParentDisplay,
    PicsDisplay
} from "@/app/components/formSubcomponents/commonClient";
import {KnownFruitableArea} from "@/app/components/formSubcomponents/knownFruitableArea";
import {SubstrateRecipeArea, SubstrateRecipeSelector} from "@/app/components/substrateRecipeClient";
import {
    ContaminationForm,
    ContamsDisplay,
    InitialContamState,
    InitialNotesState,
    IsValidContamination,
    NewContaminationForm
} from "@/app/components/formSubcomponents/contaminations";
import {GenerationInput} from "@/app/components/formSubcomponents/generationInput";
import {TopLevelImageSelector} from "@/app/components/formSubcomponents/imageSelector";
import ReaderWriterSelector, {
    WriteRfidOvcArea
} from "@/app/components/formSubcomponents/readerWriterButtons/readerSelector";
import {redirect} from "next/navigation";
import {SubstrateRecipeData} from "@/app/components/substrateRecipeServer";
import {SpeciesData} from "@/app/components/speciesServer";
import {SubspeciesData} from "@/app/components/subspeciesServer";
import {SaleArea} from "@/app/components/saleClient";
import {PcRunData, PcRunSelectorCloseable} from "@/app/components/pcRunServer";
import {BaseExternalUrl} from "@/app/components/Constants";
import {ExistingSpeciesSelector, SpeciesSubspeciesArea} from "@/app/components/speciesClient";
import {ExistingSubSpeciesSelector} from "@/app/components/subspeciesClient";
import {SubstrateBatchArea} from "@/app/components/substrateBatchClient";
import WetnessSlider from "@/app/components/formSubcomponents/utils/slider";
import {SubstrateBatchData, SubstrateBatchSelectorCloseable} from "@/app/components/substrateBatchServer";
import TestAndValidate from "@/app/components/testing/untested";
import {AclDisplay, IsValidAcl, MarshalAcl, TogglableAreaWithDepth} from "@/app/components/accessControlClient";
import {ACL} from "@/app/components/accessControlServer";
import {EntryLinkWrapper} from "@/app/components/formSubcomponents/entryLink";
import {OnViewCreatorsQuadColArea, OvcForNewFruit} from "@/app/components/formSubcomponents/ovc";

export function AssertBag(input: any): asserts input is BagData {
    if (typeof input !== 'object') {
        throw new Error('Input is not an object! Input is ' + typeof input);
    }
    // required simple keys
    let requiredSimpleKeys = new Map<string, string>([
        ['_id', 'string'],
        ['recipe', 'string'],
        ['filterSize', 'string'],
        ['creationDate', 'number'],
        ['lastUpdated', 'number'],
    ])
    for (let [key, expType] of requiredSimpleKeys) {
        if (!(key in input && typeof input[key] === expType)) {
            throw new Error('Bag assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
        }
    }

    // optional simple keys
    let optionalSimpleKeys = new Map<string, string>([
        ['pcRun', 'string'],
        ['genSpore', 'number'],
        ['genFruitOrSpore', 'number'],
        ['sealDate', 'number'],
        ['knownFruitable', 'boolean'],
        ['species', 'string'],
        ['subspecies', 'string'],
        ['innoc', 'string'],
        ['parentType', 'string'],
        ['parent', 'string'],
        ['sale', 'string'],
        ['substrateBatch', 'string'],
        ['disposed', 'number'],
        ['wetness', 'number'],
    ])
    for (let [key, expType] of optionalSimpleKeys) {
        if (!OptionalSimpleKey(key, input, expType)) {
            throw new Error('Bag assertion failure: optional key ' + key + ' was not valid');
        }
    }
    // complex required keys
    let complexRequiredKeys = new Map<string, (v: any) => boolean>([
        // ['entryType', (inp: any) => {
        //     return (typeof inp === 'string' && inp === "bag")
        // }],
    ])
    for (let [key, validator] of complexRequiredKeys) {
        if (!RequiredKey(key, input, validator)) {
            throw new Error('Bag assertion failure: required key ' + key + ' was not valid');
        }
    }
    // complex optional keys
    let complexOptionalKeys = new Map<string, (v: any) => boolean>([
        ['mostRecentImage', IsValidPicWithNotesIncoming],
        ['acl', IsValidAcl]
    ])
    for (let [key, validator] of complexOptionalKeys) {
        if (!OptionalKey(key, input, validator)) {
            throw new Error('Bag assertion failure: optional key ' + key + ' was not valid');
        }
    }
    // complex optional array keys
    let complexOptionalArrayKeys = new Map<string, (v: any) => boolean>([
        ['transfersOut', (item) => {
            return typeof item === 'string'
        }],
        ['pics', IsValidPicWithNotesIncoming],
        ['contamination', IsValidContamination],
        ['flushes', IsValidPicWithNotesIncoming],
        ['notes', IsValidNote],
    ])
    for (let [key, validator] of complexOptionalArrayKeys) {
        if (!OptionalArrayOfType(key, input, validator)) {
            throw new Error('Bag assertion failure: optional array key ' + key + ' was not valid');
        }
    }
    return
}

export default function BagDisplay(
    {
        id, readonly, data, headerLevel, isTopLevel, cookies
    }: DisplayInput) {
    try {
        AssertBag(data)
        const [initial, setInitial] = useState(data)

        const [knownFruitable, setKnownFruitable] = useState(initial.knownFruitable)
        const [sale, setSale] = useState(initial.sale)
        const [transfersOut, setTransfersOut] = useState(initial.transfersOut || [])
        const [disposed, setDisposed] = useState(initial.disposed)
        const [notes, setNotes] = useState<AllEntries<Note>>(InitialNotesState(data.notes))
        const [writeTagTo, setWriteTagTo] = useState<string | undefined>()
        // ItemsWithPics
        const [pics, setPics] = useState<SplitAllEntries<PicWithNotesForm, NewPicWithNotesForm>>(InitialPicsEntries(initial.pics))
        const [contams, setContams] = useState<SplitAllEntries<ContaminationForm, NewContaminationForm>>(InitialContamState(initial.contamination))
        const [flushes, setFlushes] = useState<SplitAllEntries<PicWithNotesForm, NewPicWithNotesForm>>(InitialPicsEntries(initial.flushes))
        const [err, setErr] = useState<string | undefined>()
        const [acl, setAcl] = useState<ACL | undefined>(initial.acl)
        //const [newFruits, setNewFruits] = useState<FruitData[]>([]) // TODO: get rid of???
        const filterSizeArea = (filterSize: string, headerLevel?: number) => {
            return <div>
                <div>{"Filter Size: " + filterSize}</div>
            </div>
        }
        const updateInitial = (updated: BagData) => {
            setInitial(updated)
            setKnownFruitable(updated.knownFruitable)
            setSale(updated.sale)
            setTransfersOut(updated.transfersOut || [])
            setDisposed(updated.disposed)
            setNotes(InitialNotesState(updated.notes))
            setPics(InitialPicsEntries(updated.pics))
            setContams(InitialContamState(updated.contamination))
            setFlushes(InitialPicsEntries(updated.flushes))
            setAcl(updated.acl)
        }
        const bagSubmit = () => {
            let body = new FormData()
            let dataObj: any = {
                knownFruitable: knownFruitable,
                sale: sale, // TODO: how/when should sales be made?
                notes: notes,
                disposed: disposed,
                writeTagTo: writeTagTo,
                acl: MarshalAcl(acl),
            }
            try {
                // Pics
                let picsInfo = resolvePicsFormData(pics)
                let newImages = picsInfo.images
                dataObj.images = picsInfo.obj
                // Contams
                let contamsInfo = resolveContamsFormData(contams)
                let newContams = contamsInfo.images
                dataObj.contams = contamsInfo.obj
                // Flushes
                let flushesInfo = resolvePicsFormData(flushes)
                let newFlushes = flushesInfo.images
                dataObj.flushes = flushesInfo.obj
                // Set data on form
                setFormData(body, dataObj)
                //body.set("data", JSON.stringify(dataObj))
                setFormImages(body, "newPic", newImages)
                setFormImages(body, "newContam", newContams)
                setFormImages(body, "newFlush", newFlushes)
            } catch (caught: any) {
                setErr(JSON.stringify(caught))
                return
            }

            SendMultipartRequest(BaseExternalUrl + "/db/update/bag/" + initial._id, cookies, body)
                .then(HandleJsonResponse)
                .then((newEntry) => {
                    AssertBag(newEntry)
                    updateInitial(newEntry)
                })
                .catch((er) => {
                    setErr(JSON.stringify(er))
                });
        }
        const ovcs: OnViewCreatorQuadCol[] = [
            OvcForNewFruit(initial._id, "bag", cookies), // TODO: test heavily
            WriteRfidOvcArea(initial._id),
        ]
        return (
            <DisplayFormWrapper entryType={"bag"}>
                <ErrorDisplay err={err} headerLevel={headerLevel}/>
                <ID id={data._id} txt={"Bag"} entryType={"bag"}/>
                <OnViewCreatorsQuadColArea OnViewCreators={ovcs} readonly={readonly}/>{/* TODO: where to put?*/}
                <MostRecentImageDisplay data={initial.mostRecentImage} headerLevel={headerLevel}/>
                <FlexedArea>
                    <FlexedSinglesGroup>
                        <SubstrateRecipeArea id={data.recipe} readonly={true} txt={"Substrate recipe: "}/>
                        <SubstrateBatchArea id={data.substrateBatch} txt={"Substrate batch: "} readonly={true}/>
                        <SubstrateRecipeArea id={initial.recipe} headerLevel={headerLevel} readonly={true}/>
                        {filterSizeArea(initial.filterSize)}
                    </FlexedSinglesGroup>
                    <FlexedSinglesGroup>
                        <DateArea pre={"PC Date: "} readonly={true} when={initial.creationDate}/>
                        <DateArea pre={"Seal Date: "} readonly={true} when={initial.sealDate}/>
                        <DateArea pre={"Last Updated: "} when={initial.lastUpdated} readonly={true}/>
                        <DisposedDisplay disposed={disposed} setDisposedOnParent={setDisposed} readonly={readonly}/>
                    </FlexedSinglesGroup>
                    <FlexedSinglesGroup>
                        <KnownFruitableArea initial={initial.knownFruitable} doSelect={setKnownFruitable}
                                            readonly={readonly}
                                            headerLevel={headerLevel}/>
                        <SaleArea sale={initial.sale} setSale={setSale} readonly={readonly}
                                  headerLevel={headerLevel} canCreateSale={true}/>
                    </FlexedSinglesGroup>
                    <FlexedSinglesGroup>
                        <GensFormDisplay gensSinceSpore={initial.genSpore}
                                         gensSinceFruitOrSpore={initial.genFruitOrSpore}/>
                        <WetnessDisplay value={initial.wetness}/>{/* TODO: FIX */}
                    </FlexedSinglesGroup>
                    <FlexedSinglesGroup>
                        <SpeciesSubspeciesArea species={initial.species} subspecies={initial.subspecies}/>
                        {/*<SpeciesSubspeciesFormArea species={initial.species} subspecies={initial.subspecies}/>*/}
                        <ParentDisplay parent={initial.parent} parentType={initial.parentType}
                                       headerLevel={headerLevel}/>
                        <InnocDisplay innoc={initial.innoc}/>
                    </FlexedSinglesGroup>
                </FlexedArea>

                <TransfersOutDisplay validTypesTo={["plate"]} thisId={initial._id} thisEntryType={"bag"}
                                     transfersOut={initial.transfersOut}
                                     allowNewTransferCreation={true}
                                     cookies={cookies}/>
                <PicsDisplay pix={initial.pics || []} readonly={readonly} updateParent={setPics}/>{/* Pics */}
                {/* Flushes */}
                <PicsDisplay pix={initial.flushes || []} readonly={readonly}
                             updateParent={setFlushes} addButtonText={"Create New Flush"}
                             sectionHeader={"Flushes: "}/>

                <ContamsDisplay initial={initial.contamination || []} updateParent={setContams}
                                readonly={readonly} headerLevel={headerLevel}/>
                <NotesFormArea readonly={readonly} initial={initial.notes} updateParent={setNotes}/>
                <TogglableAreaWithDepth startOpen={false} openTxt={"view permissions"} closeTxt={"minimize perms area"}>
                    <AclDisplay ACL={acl} readonly={readonly} updateParent={setAcl}/>
                </TogglableAreaWithDepth>
                {/* Write tag area */}
                {readonly ? null : <ReaderWriterSelector txt={"Writer to write to: "} onSelect={setWriteTagTo}
                                                         headerLevel={headerLevel}/>}
                {readonly ? null : <button className={"bottomButton greenButton"} onClick={(e) => {
                    e.stopPropagation();
                    bagSubmit()
                }}>{"Update"}</button>}
            </DisplayFormWrapper>
        )
    } catch (err) {
        return <div>{"ERROR: Bag data format incorrect: " + err}</div>
    }
}

export function WetnessDisplay({value}: { value?: number }) {
    return <TestAndValidate todos={["fix"]}>
        <div>{"Wetness: " + (value ? value + "/10" : "unknown")}</div>
    </TestAndValidate>
}

const filterSizeSelector = (setFilterSize: (f?: string) => void, filterSize?: string) => {
    const possibleFilterSizes = ["5nm", "8nm", "unknown"]
    return <div className={"centerH medGapTop"}>
        {"Filter size: "}<select className={"tailwindSelector"} value={filterSize || ""} onChange={
        (evt) => {
            setFilterSize(evt.currentTarget.value === "" ? undefined : evt.currentTarget.value)
        }
    }>
        {possibleFilterSizes.map(function (size, i) {
            return <option value={size} key={i}>{size}</option>
        })}
    </select></div>
}

export function NewBagForm({handlers, substrateBatchIn, pcRunIn}: {
    handlers: NewEntryInput<BagData>,
    pcRunIn?: PcRunData,
    substrateBatchIn?: SubstrateBatchData,
}) {
    // Required defined
    const [substrateBatch, setSubstrateBatch] = useState(substrateBatchIn)
    const [wetness, setWetness] = useState<number | undefined>(undefined)
    const [pcRun, setPcRun] = useState<PcRunData | undefined>(pcRunIn)
    const [filterSize, setFilterSize] = useState<string | undefined>()
    // Optional
    const [notes, setNotes] = useState<Note[]>([])
    const [writeTagTo, setWriteTagTo] = useState<string | undefined>()
    const [err, setErr] = useState<string | undefined>()
    const newBagSubmit = () => {
        if (pcRun === undefined) {
            setErr("PC Run cannot be undefined!");
            return
        }
        if (filterSize === undefined) {
            setErr("Filter Size cannot be undefined!");
            return
        }
        if (substrateBatch === undefined) {
            setErr("Substrate batch cannot be undefined!");
            return
        }
        let body: any = {
            pcRun: pcRun._id,
            filterSize: filterSize,
            wetness: wetness,
            substrateBatch: substrateBatch._id,
            writeTagTo: writeTagTo,
            notes: notes,
        }
        fetch(BaseExternalUrl + "/db/create/bag", {
            method: 'Post',
            body: JSON.stringify(body),
            headers: {
                credentials: 'include',
                //'Cookie': cookies,
                'Content-type': "application/json"
            },
        })
            .then(HandleJsonResponse)
            .then((newEntry) => {
                try {
                    AssertBag(newEntry)
                    handlers.onCreate && handlers.onCreate(newEntry)
                } catch (er) {
                    setErr("failed to decode response:")
                }
            })
            .catch((er) => {
                setErr(JSON.stringify(er))
            });
    }
    return (
        <NewEntryFormWrapper entryType={"bag"}>
            <ErrorDisplay err={err}/>
            <div>{"Creating Bag: "}</div>
            {substrateBatchIn !== undefined &&
                <SubstrateBatchSelectorCloseable txt={"Substrate Batch (FIXME)"} doSelect={setSubstrateBatch} allowCreation={handlers.isTopLevel} creatorInPage={false} />}
                {/*<SubstrateBatchSelector doSelect={setSubstrateBatch} allowCreate={handlers.isTopLevel}
                                         creatorInPage={false}/> TODO: Closeable?*/}
            <WetnessSlider defaultValue={5} onChange={(event: Event, value: number, activeThumb: number) => {
                setWetness(value)
            }}/>
            {pcRunIn === undefined ||
                <PcRunSelectorCloseable doSelect={setPcRun} creatorInPage={true} allowCreation={true}/>}
            {filterSizeSelector(setFilterSize, filterSize)}
            <NewEntryNotes setNotes={setNotes}/>
            {/* Write tag area */}
            <ReaderWriterSelector txt={"Writer to write to: "} onSelect={setWriteTagTo}/>
            {/* SUBMIT AREA */}
            <input type="submit" value="Submit" className={"bottomButton"} onClick={newBagSubmit} onSubmit={(e) => {
                e.preventDefault();
            }}/>
        </NewEntryFormWrapper>
    )
}

export function BagImportDisplay({headerLevel, cookies}: ImportDisplayInput) { // TODO: USE
    // Required
    const [sealDate, setSealDate] = useState(Date.now())
    const [species, setSpecies] = useState<SpeciesData | undefined>(undefined)
    const [recipe, setRecipe] = useState<SubstrateRecipeData | undefined>(undefined)
    const [filterSize, setFilterSize] = useState<string | undefined>(undefined)
    // Optional
    const [subspecies, setSubspecies] = useState<SubspeciesData | undefined>(undefined)
    const [generation, setGeneration] = useState<number | undefined>(undefined)
    const [knownFruitable, setKnownFruitable] = useState<boolean | undefined>(undefined)
    const [imageFile, setImageFile] = useState<File | undefined>(undefined)
    const [writeTagTo, setWriteTagTo] = useState<string | undefined>(undefined)
    const [err, setErr] = useState<string | undefined>(undefined)
    const submitImportBag = () => {
        const reqd = new Map<string, any>([
            ['recipe', recipe],
            ['filterSize', filterSize],
            ['species', species]
        ])
        for (let [key, val] of reqd) {
            if (val === undefined) {
                setErr(key + " must be defined!");
                return
            }
        }
        let formData = new FormData()
        let dataObj: any = {
            sealDate: sealDate,
            recipe: recipe?._id, // MUST EXIST TODO: validate on insert
            filterSize: filterSize,
            species: species?._id, // MUST EXIST TODO: validate on insert
        }

        if (subspecies !== undefined) {
            dataObj.subspecies = subspecies?._id
        }
        if (generation !== undefined) {
            dataObj.generation = generation
        }
        if (knownFruitable !== undefined) {
            dataObj.knownFruitable = knownFruitable
        }
        if (writeTagTo !== undefined) {
            dataObj.rfidWriter = writeTagTo
        }
        setFormData(formData, dataObj)
        //formData.set("data", JSON.stringify(dataObj))
        imageFile && formData.set("img", imageFile, "img")
        fetch(BaseExternalUrl + "/db/import/bag", {
            method: 'Post',
            body: formData,
            headers: {
                credentials: 'include',
                'Content-type': "multipart/form-data"
            },
        })
            .then(HandleTxtResponse) // TODO: change to json for reasons
            .then((newBagId) => {
                redirect(BaseExternalUrl + "/view/bag/" + newBagId)
            })
            .catch((err) => {
                setErr(JSON.stringify(err))
            });
    }
    return <ImportEntryFormWrapper entryType={"bag"}>
        {/* Required Fields */}
        <ErrorDisplay err={err} headerLevel={headerLevel}/>
        <DateArea pre={"Seal Date: "} when={Date.now()} updateParent={setSealDate}/>
        <SelectorWrapper current={recipe} title={"Recipe"} nameFunc={(v: SubstrateRecipeData) => v._id}>
            <SubstrateRecipeSelector doSelect={setRecipe} allowCreate={false} creatorInPage={false}/>
        </SelectorWrapper>
        {filterSizeSelector(setFilterSize, filterSize)}
        <ExistingSpeciesSelector doSelect={setSpecies}/>

        {/* Optional fields*/}
        <ExistingSubSpeciesSelector species={species?._id} doSelect={setSubspecies}/>
        <GenerationInput updateParent={setGeneration}/>
        <KnownFruitableArea doSelect={setKnownFruitable}/>

        <TopLevelImageSelector updateParent={setImageFile} buttonText={"Upload image"}/>
        <ReaderWriterSelector txt={"Write to: "} defaultOption={"none"} onSelect={setWriteTagTo}/>
        <button onClick={submitImportBag} className={"greenButton bottomButton"}>{"Submit"}</button>
    </ImportEntryFormWrapper>
}

export function BagListPageTable({data, onClick, withLink}: ListPageItems<BagData>) {
    let cols: ListTableColumn<BagData>[] = [
        NewColumn("ID", (v) => v._id),
        NewColumn("Created", (v) => {
            return NumberToDateStr(v.creationDate)
        }),
        NewColumn("Updated", (v) => {
            return NumberToDateStr(v.lastUpdated)
        }),
        NewColumn("Species", (v) => v.species || ""),
        NewColumn("Subspec.", (v) => v.subspecies || ""),
    ]
    if (withLink) {
        cols = [...cols, NewColumn("Link", (v: BagData) => {
            return <EntryLinkWrapper props={{linkId: encodeURI(v._id), entryType: "bag", openInNewTab: true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })]
    }
    // TODO: expansion for everything else????
    return <ListPageTable cols={cols} data={data} onClick={onClick}/>
}

export function BagSelectorTable({data, onClick}: ListPageItems<BagData>) {
    return <BagListPageTable data={data} onClick={onClick} withLink={true}/>
}

export function BagSelector( // TODO: USE ELSEWHERE
    {
        doSelect,
        allowCreate
    }: {
        doSelect: (val: BagData | undefined) => void,
        allowCreate?: boolean
    }) {
    const table = (items: BagData[]): JSX.Element => {
        return <BagSelectorTable data={items} onClick={doSelect}/>
    }

    return <ExistingRecentSelector entryType={"bag"} entryTypes={"bags"} doSelect={doSelect} asserter={AssertBag}
                                   table={table}>
        {allowCreate && <NewBagForm handlers={{onCreate: doSelect, isTopLevel: false}}/>}
    </ExistingRecentSelector>
}
